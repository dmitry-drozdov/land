﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using Land.Core.Specification;
using Land.Core.Lexing;
using Land.Core.Parsing.Tree;

namespace Land.Core.Parsing.LL
{
	public class Parser : BaseParser
	{
		private TableLL1 Table { get; set; }
		private FirstBuilder FirstBuilder { get; set; }

		private Stack<Node> Stack { get; set; }
		private string StackString => String.Join(" ", Stack.Select(s => GrammarObject.Developerify(s.Symbol)));

		/// <summary>
		/// Уровень вложенности относительно описанных в грамматике пар,
		/// на котором начался разбор нетерминала
		/// </summary>
		private Dictionary<Node, int> NestingLevel { get; set; }

		private Dictionary<string, Tuple<SymbolArguments, Stack<string>>> RecoveryCache { get; set; }
		private HashSet<int> PositionsWhereRecoveryStarted { get; set; }
		private Message PotentialErrorMessage { get; set; }


		public Parser(
			Grammar g,
			ILexer lexer,
			BaseNodeGenerator nodeGen = null,
			BaseNodeRetypingVisitor retypingVisitor = null) : base(g, lexer, nodeGen, retypingVisitor)
		{
			Table = new TableLL1(g);
			FirstBuilder = new FirstBuilder(g, true);

			RecoveryCache = new Dictionary<string, Tuple<SymbolArguments, Stack<string>>>();

			/// Для каждого из возможных символов для восстановления кешируем дополнительную информацию
			foreach (var smb in GrammarObject.Options.GetSymbols(ParsingOption.GROUP_NAME, ParsingOption.RECOVERY))
			{
				var stack = new Stack<string>();
				stack.Push(smb);
				SymbolArguments anyArgs = null;

				/// Накапливаем символы, которые придётся положить на стек при восстановлении на smb
				while (true)
				{
					var alternative = Table[stack.Peek(), Grammar.ANY_TOKEN_NAME][0];
					stack.Pop();

					for (var i = alternative.Count - 1; i >= 0; --i)
						stack.Push(alternative[i]);

					if (alternative[0].Symbol == Grammar.ANY_TOKEN_NAME)
					{
						anyArgs = alternative[0].Arguments;
						stack.Pop();
						break;
					}
				}

				/// В кеш помещаем цепочку символов и опции Any
				RecoveryCache[smb] = new Tuple<SymbolArguments, Stack<string>>(anyArgs, stack);
			}
		}

		/// <summary>
		/// LL(1) разбор
		/// </summary>
		/// <returns>
		/// Корень дерева разбора
		/// </returns>
		protected override (Node, Durations) ParsingAlgorithm(string text)
		{
			/// Контроль вложенностей пар
			NestingLevel = new Dictionary<Node, int>();
			PositionsWhereRecoveryStarted = new HashSet<int>();

			/// Готовим лексер и стеки
			LexingStream = new ComplexTokenStream(GrammarObject, Lexer, text, Log);
			Stack = new Stack<Node>();
			/// Кладём на стек стартовый символ
			var root = NodeGenerator.Generate(GrammarObject.StartSymbol);
			Stack.Push(NodeGenerator.Generate(Grammar.EOF_TOKEN_NAME));
			Stack.Push(root);

			var d = new Durations();

			/// Читаем первую лексему из входного потока
			var token = LexingStream.GetNextToken();

			/// Пока не прошли полностью правило для стартового символа
			while (Stack.Count > 0)
			{
				if (token.Type == Grammar.ERROR_TOKEN_TYPE)
					break;

				var stackTop = Stack.Peek();

				if (EnableTracing)
				{
					Log.Add(Message.Trace(
						$"Текущий токен: {this.Developerify(token)}\t |\t Стек: {StackString}",
						LexingStream.CurrentToken.Location.Start
					));
				}

				/// Если символ на вершине стека совпадает с текущим токеном
				if (stackTop.Symbol == token.Name)
				{
					if (token.Type == Grammar.ANY_TOKEN_TYPE)
					{
						token = SkipAny(NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME), true);
					}
					else
					{
						var node = Stack.Pop();
						node.SetLocation(token.Location.Start, token.Location.End);
						node.SetValue(token.Text);

						token = LexingStream.GetNextToken();
					}

					continue;
				}

				/// Если на вершине стека нетерминал, выбираем альтернативу по таблице
				if (GrammarObject[stackTop.Symbol] is NonterminalSymbol)
				{
					var alternatives = Table[stackTop.Symbol, token.Name];

					if (alternatives.Count > 0)
					{
						if (token.Type == Grammar.ANY_TOKEN_TYPE)
						{
							/// Поддерживаем свойство immediate error detection для Any
							var runtimeFirst = Stack.Select(e => e.Symbol).ToList();

							if (FirstBuilder.First(runtimeFirst).Contains(Grammar.ANY_TOKEN_NAME))
								token = SkipAny(NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME), true);
							else
							{
								Log.Add(PotentialErrorMessage = Message.Trace(
									$"Неожиданный токен {this.Developerify(LexingStream.CurrentToken)}, ожидались {String.Join(", ", runtimeFirst.Select(t => this.Developerify(t)))}",
									token.Location.Start,
									addInfo: new Dictionary<MessageAddInfoKey, object>
									{
										{ MessageAddInfoKey.UnexpectedToken, LexingStream.CurrentToken.Name },
										{ MessageAddInfoKey.UnexpectedLexeme, LexingStream.CurrentToken.Text },
										{ MessageAddInfoKey.ExpectedTokens, runtimeFirst.ToList() }
									}
								));

								token = ErrorRecovery();
							}
						}
						else
						{
							ApplyAlternative(alternatives[0]);
						}

						continue;
					}
				}

				/// Если не смогли ни сопоставить текущий токен с терминалом на вершине стека,
				/// ни найти ветку правила для нетерминала на вершине стека
				if (token.Type == Grammar.ANY_TOKEN_TYPE)
				{
					Log.Add(PotentialErrorMessage = Message.Trace(
						GrammarObject.Tokens.ContainsKey(stackTop.Symbol) ?
							$"Неожиданный токен {this.Developerify(LexingStream.CurrentToken)}, ожидался {this.Developerify(stackTop.Symbol)}" :
							$"Неожиданный токен {this.Developerify(LexingStream.CurrentToken)}, ожидались {String.Join(", ", Table[stackTop.Symbol].Where(t => t.Value.Count > 0).Select(t => this.Developerify(t.Key)))}",
						LexingStream.CurrentToken.Location.Start,
						addInfo: new Dictionary<MessageAddInfoKey, object>
						{
							{ MessageAddInfoKey.UnexpectedToken, LexingStream.CurrentToken.Name },
							{ MessageAddInfoKey.UnexpectedLexeme, LexingStream.CurrentToken.Text },
							{
								MessageAddInfoKey.ExpectedTokens,
								GrammarObject.Tokens.ContainsKey(stackTop.Symbol)
									? new List<string> { stackTop.Symbol }
									: Table[stackTop.Symbol].Where(t => t.Value.Count > 0).Select(e => e.Key).ToList()
							}
						}
					));

					token = ErrorRecovery();
				}
				/// Если непонятно, что делать с текущим токеном, и он конкретный
				/// (не Any), заменяем его на Any
				else
				{
					/// Если встретился неожиданный токен, но он в списке пропускаемых
					if (GrammarObject.Options.IsSet(ParsingOption.GROUP_NAME, ParsingOption.SKIP, token.Name))
					{
						token = LexingStream.GetNextToken();
					}
					else
					{
						if (EnableTracing)
						{
							Log.Add(Message.Trace(
								$"Попытка трактовать текущий токен как начало участка, соответствующего Any",
								token.Location.Start
							));
						}

						token = Lexer.CreateToken(Grammar.ANY_TOKEN_NAME, Grammar.ANY_TOKEN_TYPE);
					}
				}
			}

			root = TreePostProcessing(root);

			return (root, d);
		}

		private HashSet<string> GetStopTokens(SymbolArguments args, IEnumerable<string> followSequence)
		{
			/// Если с Any не связана последовательность стоп-символов
			if (!args.AnyArguments.ContainsKey(AnyArgument.Except))
			{
				/// Определяем множество токенов, которые могут идти после Any
				var tokensAfterText = FirstBuilder.First(followSequence.ToList());
				/// Само Any во входном потоке нам и так не встретится, а вывод сообщения об ошибке будет красивее
				tokensAfterText.Remove(Grammar.ANY_TOKEN_NAME);

				/// Если указаны токены, которые нужно однозначно включать в Any
				if (args.AnyArguments.ContainsKey(AnyArgument.Include))
				{
					tokensAfterText.ExceptWith(args.AnyArguments[AnyArgument.Include]);
				}

				return tokensAfterText;
			}
			else
			{
				return args.AnyArguments[AnyArgument.Except];
			}
		}

		/// <summary>
		/// Пропуск токенов в позиции, задаваемой символом Any
		/// </summary>
		/// <returns>
		/// Токен, найденный сразу после символа Any
		/// </returns>
		private IToken SkipAny(Node anyNode, bool enableRecovery)
		{
			var nestingCopy = LexingStream.GetPairsState();
			var tokenIndex = LexingStream.CurrentIndex;
			var token = LexingStream.CurrentToken;

			var stackTop = Stack.Peek();
			var d = new Durations();

			if (EnableTracing)
				Log.Add(Message.Trace(
					$"Инициирован пропуск Any\t |\t Стек: {StackString}",
					token.Location.Start
				));

			/// Пока по Any нужно раскрывать очередной нетерминал
			while (GrammarObject[stackTop.Symbol] is NonterminalSymbol)
			{
				ApplyAlternative(Table[stackTop.Symbol, Grammar.ANY_TOKEN_NAME][0]);
				stackTop = Stack.Peek();
			}

			/// В итоге первым терминалом, который окажется на стеке, должен быть Any
			/// Подменяем свежесгенерированный узел для Any на переданный извне
			anyNode.Options = stackTop.Options.Clone();
			anyNode.Arguments = stackTop.Arguments.Clone();

			/// Проверяем, не происходит ли восстановление в действительно некорректной программе
			if (anyNode.Arguments.Contains(AnyArgument.Error)
				&& PotentialErrorMessage != null)
			{
				PotentialErrorMessage.Type = MessageType.Error;
			}

			var anyIndex = stackTop.Parent.Children.IndexOf(stackTop);
			stackTop.Parent.ReplaceChild(anyNode, anyIndex);
			Stack.Pop();

			if (EnableTracing)
				Log.Add(Message.Trace(
					$"Поиск окончания последовательности, соответствующей Any\t |\t Стек: {StackString}",
					token.Location.Start
				));

			var stopTokens = GetStopTokens(anyNode.Arguments, Stack.Select(n => n.Symbol));
			var ignorePairs = anyNode.Arguments.Contains(AnyArgument.IgnorePairs);

			/// Если Any непустой (текущий токен - это не токен,
			/// который может идти после Any)
			if (!stopTokens.Contains(token.Name))
			{
				/// Проверка на случай, если допропускаем текст в процессе восстановления
				if (anyNode.Location == null)
					anyNode.SetLocation(token.Location.Start, token.Location.End);

				/// Смещение для участка, подобранного как текст
				var endLocation = token.Location.End;
				var anyLevel = LexingStream.GetPairsCount();

				while (!stopTokens.Contains(token.Name)
					&& (ignorePairs || LexingStream.CurrentTokenDirection != Direction.Up)
					&& !anyNode.Arguments.Contains(AnyArgument.Avoid, token.Name)
					&& token.Name != Grammar.EOF_TOKEN_NAME
					&& token.Type != Grammar.ERROR_TOKEN_TYPE)
				{
					anyNode.Value.Add(token.Text);
					endLocation = token.Location.End;

					if (ignorePairs)
					{
						token = LexingStream.GetNextToken();
					}
					else
					{
						token = LexingStream.GetNextToken(anyLevel, out List<IToken> skippedBuffer);

						/// Если при пропуске до токена на том же уровне
						/// пропустили токены с более глубокой вложенностью
						if (skippedBuffer.Count > 0)
						{
							anyNode.Value.AddRange(skippedBuffer.Select(t => t.Text));
							endLocation = skippedBuffer.Last().Location.End;
						}
					}
				}

				anyNode.SetLocation(anyNode.Location.Start, endLocation);

				if (token.Type == Grammar.ERROR_TOKEN_TYPE)
					return token;

				if (!stopTokens.Contains(token.Name))
				{
					var message = Message.Trace(
						$"Ошибка при пропуске {Grammar.ANY_TOKEN_NAME}: неожиданный токен {GrammarObject.Developerify(token.Name)}, ожидались {String.Join(", ", stopTokens.Select(t => this.Developerify(t)))}",
						token.Location.Start,
						addInfo: new Dictionary<MessageAddInfoKey, object>
						{
							{ MessageAddInfoKey.UnexpectedToken, token.Name },
							{ MessageAddInfoKey.UnexpectedLexeme, token.Text },
							{ MessageAddInfoKey.ExpectedTokens, stopTokens.ToList() }
						}
					);

					Log.Add(message);

					if (enableRecovery)
					{
						++Statistics.RecoveryTimesAny;
						Statistics.LongestRollback =
							Math.Max(Statistics.LongestRollback, LexingStream.CurrentIndex - tokenIndex);

						LexingStream.MoveTo(tokenIndex, nestingCopy);
						anyNode.Reset();
						///	Возвращаем узел обратно на стек
						Stack.Push(anyNode);

						PotentialErrorMessage = message;

						return ErrorRecovery(stopTokens,
							anyNode.Arguments.Contains(AnyArgument.Avoid, token.Name) ? token.Name : null);
					}
					else
					{
						PotentialErrorMessage.Type = MessageType.Error;

						return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
					}
				}
			}

			return token;
		}

		private void ApplyAlternative(Alternative alternativeToApply)
		{
			var stackTop = Stack.Pop();

			if (!String.IsNullOrEmpty(alternativeToApply.Alias))
				stackTop.Alias = alternativeToApply.Alias;

			NestingLevel[stackTop] = LexingStream.GetPairsCount();

			for (var i = alternativeToApply.Count - 1; i >= 0; --i)
			{
				var newNode = NodeGenerator.Generate(
					alternativeToApply[i].Symbol,
					alternativeToApply[i].Options.Clone(),
					alternativeToApply[i].Arguments.Clone()
				);

				stackTop.AddFirstChild(newNode);
				Stack.Push(newNode);
			}
		}

		private IToken ErrorRecovery(HashSet<string> stopTokens = null, string avoidedToken = null)
		{
			var d = new Durations();
			// Если восстановление от ошибок отключено на уровне грамматики
			if (!GrammarObject.Options.IsRecoveryEnabled())
			{
				PotentialErrorMessage.Type = MessageType.Error;
				return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
			}

			// Если в текущей позиции уже запускалось восстановление
			if (!PositionsWhereRecoveryStarted.Add(LexingStream.CurrentIndex))
			{
				PotentialErrorMessage.Type = MessageType.Error;

				Log.Add(Message.Error(
					$"Возобновление разбора невозможно: восстановление в позиции токена {this.Developerify(LexingStream.CurrentToken)} уже проводилось",
					LexingStream.CurrentToken.Location.Start
				));

				return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
			}

			Log.Add(Message.Trace(
				$"Процесс восстановления запущен в позиции токена {this.Developerify(LexingStream.CurrentToken)}",
				LexingStream.CurrentToken.Location.Start
			));

			var recoveryStartTime = DateTime.UtcNow;

			// То, что мы хотели разобрать, и не смогли
			var currentNode = Stack.Pop();

			// Поднимаемся по уже построенной части дерева, пока не встретим 
			// пригодный для восстановления нетерминал. 
			do
			{
				if (currentNode.Parent != null)
				{
					var childIndex = currentNode.Parent.Children.IndexOf(currentNode);

					for (var i = 0; i < currentNode.Parent.Children.Count - childIndex - 1; ++i)
						Stack.Pop();
				}

				// Переходим к родителю
				currentNode = currentNode.Parent;
			}
			// Ищем дальше, если
			while (currentNode != null && (
				// текущий символ не входит в список тех, на которых можно восстановиться, или
				!GrammarObject.Options.IsSet(ParsingOption.GROUP_NAME, ParsingOption.RECOVERY, currentNode.Symbol) ||
				// при разборе соответствующей сущности уже пошли по Any-ветке
				ParsedStartsWithAny(currentNode) ||
				// ошибка произошла на таком же Any
				IsUnsafeAny(stopTokens, avoidedToken, currentNode)
			));

			if (currentNode != null)
			{
				List<IToken> skippedBuffer;

				if (LexingStream.GetPairsCount() != NestingLevel[currentNode])
				{
					var currentToken = LexingStream.CurrentToken;

					// Пропускаем токены, пока не поднимемся на тот же уровень вложенности, 
					// на котором раскрывали нетерминал
					var nonterminalLevelToken = LexingStream.GetNextToken(NestingLevel[currentNode],  out skippedBuffer);

					if (nonterminalLevelToken.Type != Grammar.ERROR_TOKEN_TYPE)
					{
						skippedBuffer.Insert(0, currentToken);
					}
					else
					{
						PotentialErrorMessage.Type = MessageType.Error;
						return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
					}
				}
				else
				{
					skippedBuffer = new List<IToken>();
				}

				var anyNode = NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME);

				anyNode.Value = currentNode.GetValue();
				anyNode.Value.AddRange(skippedBuffer.Select(t => t.Text));

				if (currentNode.Location != null)
					anyNode.SetLocation(currentNode.Location.Start, currentNode.Location.End);

				currentNode.ResetChildren();
				Stack.Push(currentNode);

				if (skippedBuffer.Count > 0)
				{
					anyNode.SetLocation(
						anyNode.Location?.Start ?? skippedBuffer[0].Location.Start,
						skippedBuffer.Last().Location.End
					);
				}

				Log.Add(Message.Trace(
					$"Найдено предполагаемое начало {Grammar.ANY_TOKEN_NAME}",
					anyNode.Location?.Start ?? LexingStream.CurrentToken.Location.Start
				));

				Log.Add(Message.Trace(
					$"Попытка продолжить разбор на нетерминале {GrammarObject.Developerify(currentNode.Symbol)} в позиции токена {this.Developerify(LexingStream.CurrentToken)}",
					LexingStream.CurrentToken.Location.Start
				));

				// Пытаемся пропустить Any в этом месте
				var token = SkipAny(anyNode, false);

				// Если Any успешно пропустили и возобновили разбор,
				// возвращаем токен, с которого разбор продолжается
				if (token.Type != Grammar.ERROR_TOKEN_TYPE)
				{
					Log.Add(Message.Trace(
						$"Произведено восстановление на уровне {GrammarObject.Developerify(currentNode.Symbol)}, разбор продолжен с токена {this.Developerify(token)}",
						token.Location.Start
					));

					Statistics.RecoveryTimes += 1;
					Statistics.RecoveryTimeSpent += DateTime.UtcNow - recoveryStartTime;

					return token;
				}
			}

			PotentialErrorMessage.Type = MessageType.Error;
			return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
		}

		private bool ParsedStartsWithAny(Node subtree)
		{
			while (subtree.Symbol != Grammar.ANY_TOKEN_NAME
				&& subtree.Children.Count > 0)
				subtree = subtree.Children[0];

			return subtree.Symbol == Grammar.ANY_TOKEN_NAME;
		}

		private bool IsUnsafeAny(HashSet<string> oldStopTokens, string avoidedToken, Node currentNode)
		{
			return oldStopTokens != null && LexingStream.GetPairsCount() == NestingLevel[currentNode] && (
				RecoveryCache[currentNode.Symbol].Item1.Contains(AnyArgument.Avoid, LexingStream.CurrentToken.Name)
				|| GetStopTokens(RecoveryCache[currentNode.Symbol].Item1,
					RecoveryCache[currentNode.Symbol].Item2.Concat(Stack.Select(n => n.Symbol))).Except(oldStopTokens).Count() == 0
				&& (avoidedToken == null || RecoveryCache[currentNode.Symbol].Item1.Contains(AnyArgument.Avoid, avoidedToken))
			);
		}
	}
}
