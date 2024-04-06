﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Land.Core.Specification;
using Land.Core.Lexing;
using Land.Core.Parsing.Tree;
using System.Diagnostics;
using System.Runtime.CompilerServices;

namespace Land.Core.Parsing.LR
{
	public class Parser : BaseParser
	{
		private TableLR1 Table { get; set; }

		private Stack<int> StatesStack { get; set; } = new Stack<int>();
		private Stack<Node> SymbolsStack { get; set; } = new Stack<Node>();
		private Stack<int> NestingStack { get; set; }

		private HashSet<int> PositionsWhereRecoveryStarted { get; set; }
		private Message PotentialErrorMessage { get; set; }


		public Parser(
			Grammar g,
			ILexer lexer,
			BaseNodeGenerator nodeGen = null,
			BaseNodeRetypingVisitor retypingVisitor = null) : base(g, lexer, nodeGen, retypingVisitor)
		{
			Table = new TableLR1(g);
		}

		protected override (Node, Durations) ParsingAlgorithm(string text)
		{
			Node root = null;

			//var watcher = Stopwatch.StartNew();

			EnableTracing = false; //debug

			//d.Start();

			/// Множество индексов токенов, на которых запускалось восстановление
			PositionsWhereRecoveryStarted = new HashSet<int>();
			/// Создаём стек для уровней вложенности пар
			NestingStack = new Stack<int>();
			/// Готовим лексер
			LexingStream = new ComplexTokenStream(GrammarObject, Lexer, text, Log);
			/// Читаем первую лексему из входного потока
			var token = LexingStream.GetNextToken();
			/// Создаём стек
			StatesStack = new Stack<int>();
			SymbolsStack = new Stack<Node>();
			StatesStack.Push(0);
			NestingStack.Push(0);

			//d.Stop("init");


			while (true)
			{
				//d.Start();
				if (token.Type == Grammar.ERROR_TOKEN_TYPE)
					break;


				var currentState = StatesStack.Peek();
				//d.Stop("PeekState");

				if (EnableTracing && token.Type != Grammar.ERROR_TOKEN_TYPE && token.Type != Grammar.ANY_TOKEN_TYPE)
					Log.Add(Message.Trace(
						$"Текущий токен: {this.Developerify(token)} | Стек: TODO Stack.ToString(GrammarObject)",
						token.Location.Start
					));

				//d.Start();
				//d.Stop("cnt");

				var action = Table[currentState, token.Name];

				if (action != null)
				{
					if (token.Type == Grammar.ANY_TOKEN_TYPE)
					{
						//d.Start();
						token = SkipAny(NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME), true);
						//d.Stop("SkipAny");

						/// Если при пропуске текста произошла ошибка, прерываем разбор
						if (token.Type == Grammar.ERROR_TOKEN_TYPE)
							break;
						else
							continue;
					}

					//d.Start();

					//d.Stop("GetAction");

					//d.Start();
					/// Если нужно произвести перенос
					if (action.ActionType == 0)
					{
						var tokenNode = NodeGenerator.Generate(token.Name);
						tokenNode.SetValue(token.Text);
						tokenNode.SetLocation(token.Location.Start, token.Location.End);

						/// Вносим в стек новое состояние
						SymbolsStack.Push(tokenNode);
						StatesStack.Push(action.TargetItemIndex);
						NestingStack.Push(LexingStream.GetPairsCount());

						if (EnableTracing)
						{
							Log.Add(Message.Trace(
								$"Перенос",
								token.Location.Start
							));
						}

						token = LexingStream.GetNextToken();
						//d.Stop("ShiftAction");
					}
					/// Если нужно произвести свёртку
					else if (action.ActionType == 1)
					{
						var parentNode = NodeGenerator.Generate(action.ReductionAlternative.NonterminalSymbolName);

						/// Снимаем со стека символы ветки, по которой нужно произвести свёртку
						for (var i = 0; i < action.ReductionAlternative.Count; ++i)
						{
							parentNode.AddFirstChild(SymbolsStack.Peek());
							SymbolsStack.Pop();
							StatesStack.Pop();
							NestingStack.Pop();
						}
						currentState = StatesStack.Peek();

						/// Кладём на стек состояние, в которое нужно произвести переход
						SymbolsStack.Push(parentNode);
						StatesStack.Push(Table.Transitions[currentState][action.ReductionAlternative.NonterminalSymbolName]);

						NestingStack.Push(LexingStream.GetPairsCount());

						if (EnableTracing)
						{
							Log.Add(Message.Trace(
								$"Свёртка по правилу {GrammarObject.Developerify(action.ReductionAlternative)} -> {GrammarObject.Developerify(action.ReductionAlternative.NonterminalSymbolName)}",
								token.Location.Start
							));
						}
						//d.Stop("ReduceAction");
						continue;
					}
					else if (action.ActionType == 2)
					{
						root = SymbolsStack.Peek();
						//d.Stop("PeekSymbol");
						break;
					}
				}
				else if (token.Type == Grammar.ANY_TOKEN_TYPE)
				{
					//d.Start();

					Log.Add(PotentialErrorMessage = Message.Trace(
						$"Неожиданный символ {this.Developerify(LexingStream.CurrentToken)} для состояния{Environment.NewLine}\t\t" + Table.ToString(StatesStack.Peek(), null, "\t\t"),
						LexingStream.CurrentToken.Location.Start,
						addInfo: new Dictionary<MessageAddInfoKey, object>
						{
							{
								MessageAddInfoKey.UnexpectedToken,
								LexingStream.CurrentToken.Name
							},
							{
								MessageAddInfoKey.UnexpectedLexeme,
								LexingStream.CurrentToken.Text
							},
							{
								MessageAddInfoKey.ExpectedTokens,
								Table.Items[StatesStack.Peek()].Markers
									.Where(i=>i.Lookahead != null).Select(e => e.Lookahead).ToList()
							}
						}
					));


					token = ErrorRecovery();
					//d.Stop("ErrorRecovery");
				}
				else
				{
					//d.Start();
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
					//d.Stop("Unexpected token");
				}
			}

			//d.Start();
			if (root != null)
				root = TreePostProcessing(root);
			//d.Stop("TreePostProcessing");

			//d.Add("ParsingAlgorithm", watcher);

			return (root, null);
		}


		[MethodImpl(MethodImplOptions.AggressiveInlining)]
		private IToken SkipAny(Node anyNode, bool enableRecovery)
		{
			var nestingCopy = LexingStream.GetPairsState();
			var token = LexingStream.CurrentToken;
			var tokenIndex = LexingStream.CurrentIndex;
			var peekState = StatesStack.Peek();
			var action = Table[peekState, Grammar.ANY_TOKEN_NAME];
			var conflict = Table.Conflict(peekState, Grammar.ANY_TOKEN_NAME);

			if (EnableTracing)
			{
				/*Log.Add(Message.Trace(
					$"Инициирован пропуск Any | Стек: {Stack.ToString(GrammarObject)} | Состояние: {Environment.NewLine}\t\t"
						+ Table.ToString(Stack.PeekState(), null, "\t\t"),
					token.Location.Start
				));*/
			}

			/// Пока по Any нужно производить свёртки (ячейка таблицы непуста и нет конфликтов)

			while (action != null && action.ActionType == 1 && !conflict)
			{
				var parentNode = NodeGenerator.Generate(action.ReductionAlternative.NonterminalSymbolName);

				/// Снимаем со стека символы ветки, по которой нужно произвести свёртку
				for (var i = 0; i < action.ReductionAlternative.Count; ++i)
				{
					parentNode.AddFirstChild(SymbolsStack.Peek());
					SymbolsStack.Pop();
					StatesStack.Pop();
					NestingStack.Pop();
				}

				/// Кладём на стек состояние, в которое нужно произвести переход
				var state = Table.Transitions[StatesStack.Peek()][action.ReductionAlternative.NonterminalSymbolName];
				StatesStack.Push(state);
				SymbolsStack.Push(parentNode);
				NestingStack.Push(LexingStream.GetPairsCount());

				action = Table[state, Grammar.ANY_TOKEN_NAME];
				conflict = Table.Conflict(state, Grammar.ANY_TOKEN_NAME);
			}

			/// Берём опции из нужного вхождения Any
			var anyEntry = Table.Items[StatesStack.Peek()].AnyEntry;

			anyNode.Options = anyEntry.Options;
			anyNode.Arguments = anyEntry.Arguments;

			/// Проверяем, не происходит ли восстановление в действительно некорректной программе
			if (anyNode.Arguments.Contains(AnyArgument.Error)
				&& PotentialErrorMessage != null)
			{
				PotentialErrorMessage.Type = MessageType.Error;
			}

			/// Производим перенос
			/// Вносим в стек новое состояние
			StatesStack.Push(action.TargetItemIndex);
			SymbolsStack.Push(anyNode);
			NestingStack.Push(LexingStream.GetPairsCount());


			if (EnableTracing)
			{
				/*Log.Add(Message.Trace(
					$"Поиск окончания последовательности, соответствующей Any | Стек: {Stack.ToString(GrammarObject)} | Состояние: {Environment.NewLine}\t\t"
						+ Table.ToString(Stack.PeekState(), null, "\t\t"),
					token.Location.Start
				));*/
			}

			var stopTokens = GetStopTokens(anyNode.Arguments, StatesStack.Peek());
			var ignorePairs = anyNode.Arguments.Contains(AnyArgument.IgnorePairs);

			var startLocation = anyNode.Location?.Start
				?? token.Location.Start;
			var endLocation = anyNode.Location?.End;
			var anyLevel = LexingStream.GetPairsCount();


			/// Пропускаем токены, пока не найдём тот, для которого
			/// в текущем состоянии нужно выполнить перенос или свёртку
			while (!stopTokens.Contains(token.Name)
				&& (ignorePairs || LexingStream.CurrentTokenDirection != Direction.Up)
				&& !anyNode.Arguments.Contains(AnyArgument.Avoid, token.Name)
				&& token.Type != Grammar.EOF_TOKEN_TYPE
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

					//d.Start();
					if (skippedBuffer.Count > 0)
					{
						anyNode.Value.AddRange(skippedBuffer.Select(t => t.Text));
						endLocation = skippedBuffer.Last().Location.End;
					}
					//d.Stop("AddRange");
				}
			}


			if (endLocation != null)
			{
				anyNode.SetLocation(startLocation, endLocation);
			}

			if (token.Type == Grammar.ERROR_TOKEN_TYPE)
			{
				return token;
			}

			/// Если дошли до конца входной строки, и это было не по плану
			if (!stopTokens.Contains(token.Name))
			{
				if (enableRecovery)
				{
					var message = Message.Trace(
						$"Ошибка при пропуске {Grammar.ANY_TOKEN_NAME}: неожиданный токен {this.Developerify(token)}, ожидались {String.Join(", ", stopTokens.Select(t => this.Developerify(t)))}",
						token.Location.Start,
						addInfo: new Dictionary<MessageAddInfoKey, object>
						{
							{ MessageAddInfoKey.UnexpectedToken, token.Name },
							{ MessageAddInfoKey.UnexpectedLexeme, token.Text },
							{ MessageAddInfoKey.ExpectedTokens, stopTokens.ToList() }
						}
					);

					Log.Add(message);

					if (GrammarObject.Options.IsRecoveryEnabled())
					{
						++Statistics.RecoveryTimesAny;
						Statistics.LongestRollback =
							Math.Max(Statistics.LongestRollback, LexingStream.CurrentIndex - tokenIndex);

						LexingStream.MoveTo(tokenIndex, nestingCopy);

						PotentialErrorMessage = message;

						return ErrorRecovery(stopTokens,
							anyNode.Arguments.Contains(AnyArgument.Avoid, token.Name) ? token.Name : null);
					}
					else
					{
						message.Type = MessageType.Error;
						return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
					}
				}
				else
				{
					Log.Add(Message.Trace(
						$"Ошибка при пропуске {Grammar.ANY_TOKEN_NAME} в процессе восстановления: неожиданный токен {this.Developerify(token)}, ожидались {String.Join(", ", stopTokens.Select(t => this.Developerify(t)))}",
						token.Location.Start,
						addInfo: new Dictionary<MessageAddInfoKey, object>
						{
							{ MessageAddInfoKey.UnexpectedToken, token.Name },
							{ MessageAddInfoKey.UnexpectedLexeme, token.Text },
							{ MessageAddInfoKey.ExpectedTokens, stopTokens.ToList() }
						}
					));

					PotentialErrorMessage.Type = MessageType.Error;
					return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
				}
			}

			return token;
		}

		public HashSet<string> GetStopTokens(SymbolArguments args, int state)
		{
			HashSet<string> stopTokens;
			if (args.Contains(AnyArgument.Except))
			{
				stopTokens = args.AnyArguments[AnyArgument.Except];
			}
			else if (args.Contains(AnyArgument.Include))
			{
				stopTokens = Table.GetExpectedTokens(state);
				stopTokens.ExceptWith(args.AnyArguments[AnyArgument.Include]);
			}
			else
			{
				stopTokens = Table.GetExpectedTokens(state);
			}
			stopTokens.Remove(Grammar.ANY_TOKEN_NAME);

			return stopTokens;
		}

		public class PathFragment
		{
			public Alternative Alt { get; set; }
			public int Pos { get; set; }

			public override bool Equals(object obj)
			{
				return obj is PathFragment pf
					&& Alt.Equals(pf.Alt) && Pos == pf.Pos;
			}

			public override int GetHashCode()
			{
				return Alt.GetHashCode();
			}
		}

		private IToken ErrorRecovery(HashSet<string> stopTokens = null, string avoidedToken = null)
		{
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

			// Координаты начала и конца уже сопоставленной последовательности токенов,
			// которую теперь будем интерпретировать как Any
			PointLocation startLocation = null;
			PointLocation endLocation = null;

			// Лексемы, соответствующие этой последовательности токенов
			var value = new List<string>();

			var previouslyMatched = (Node)null;
			var derivationProds = new HashSet<PathFragment>();
			var initialDerivationProds = new HashSet<PathFragment>();

			// Снимаем со стека состояния до тех пор, пока не находим состояние,
			// в котором есть пункт A -> * Any ...
			do
			{
				if (SymbolsStack.Count > 0)
				{
					if (SymbolsStack.Peek().Location != null)
					{
						startLocation = SymbolsStack.Peek().Location.Start;
						if (endLocation == null)
						{
							endLocation = SymbolsStack.Peek().Location.End;
						}
					}

					value = SymbolsStack.Peek().GetValue()
						.Concat(value).ToList();

					// Запоминаем снятый со стека символ - это то, что было успешно распознано
					previouslyMatched = SymbolsStack.Peek();


				}

				SymbolsStack.Pop();
				StatesStack.Pop();
				NestingStack.Pop();

				if (StatesStack.Count > 0)
				{
					// Выбираем пункты, продукции которых потенциально могут участвовать
					// в выводе текущего префикса из стартового символа
					initialDerivationProds = new HashSet<PathFragment>(
						Table.Items[StatesStack.Peek()].Markers
							.Where
							(i =>
								// Точка должна стоять перед символом, только что снятым со стека
								i.Next == previouslyMatched.Symbol &&
								// Если это не первая выборка, на предыдущем шаге в выборке должен был быть пункт
								// с той же альтернативой, но точкой на один символ дальше
								(derivationProds.Count == 0 || derivationProds.Any(p => p.Alt.Equals(i.Alternative) && p.Pos == i.Position + 1))
							)
							.Select(i => new PathFragment { Alt = i.Alternative, Pos = i.Position })
					);

					derivationProds = new HashSet<PathFragment>(initialDerivationProds);

					var oldCount = 0;

					while (oldCount != derivationProds.Count)
					{
						oldCount = derivationProds.Count;

						// Добавляем к списку пункты, порождающие уже добавленные пункты
						derivationProds.UnionWith(Table.Items[StatesStack.Peek()].Markers
							.Where(i => derivationProds.Any(p => p.Pos == 0 && p.Alt.NonterminalSymbolName == i.Next))
							.Select(i => new PathFragment { Alt = i.Alternative, Pos = i.Position })
						);
					}
				}
			}
			while (StatesStack.Count > 0 && (derivationProds.Count == initialDerivationProds.Count
				|| derivationProds.Except(initialDerivationProds).All(p => !GrammarObject.Options.IsSet(ParsingOption.GROUP_NAME, ParsingOption.RECOVERY, p.Alt[p.Pos]))
				|| StartsWithAny(previouslyMatched)
				|| IsUnsafeAny(stopTokens, avoidedToken))
			);

			if (StatesStack.Count > 0)
			{
				if (LexingStream.GetPairsCount() != NestingStack.Peek())
				{
					var skippedBuffer = new List<IToken>();

					// Запоминаем токен, на котором произошла ошибка
					var currentToken = LexingStream.CurrentToken;

					// Пропускаем токены, пока не поднимемся на тот же уровень вложенности, 
					// на котором раскрывали нетерминал
					var nonterminalLevelToken = LexingStream.GetNextToken(NestingStack.Peek(), out skippedBuffer);

					if (nonterminalLevelToken.Type != Grammar.ERROR_TOKEN_TYPE)
					{
						skippedBuffer.Insert(0, currentToken);

						value.AddRange(skippedBuffer.Select(t => t.Text));
						endLocation = skippedBuffer.Last().Location.End;
					}
					else
					{
						PotentialErrorMessage.Type = MessageType.Error;
						return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
					}
				}

				// Пытаемся пропустить Any в этом месте,
				// Any захватывает участок с начала последнего 
				// снятого со стека символа до места восстановления
				var anyNode = NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME);
				if (startLocation != null)
					anyNode.SetLocation(startLocation, endLocation);
				anyNode.Value = value.ToList();

				Log.Add(Message.Trace(
					$"Найдено предполагаемое начало {Grammar.ANY_TOKEN_NAME}",
					anyNode.Location?.Start ?? LexingStream.CurrentToken.Location.Start
				));

				/*Log.Add(Message.Trace(
					$"Попытка продолжить разбор в состоянии {Environment.NewLine}\t\t{Table.ToString(Stack.PeekState(), null, "\t\t")}\tв позиции токена {this.Developerify(LexingStream.CurrentToken)}",
					LexingStream.CurrentToken.Location.Start
				));*/

				var token = SkipAny(anyNode, false);

				// Если Any успешно пропустили и возобновили разбор,
				// возвращаем токен, с которого разбор продолжается
				if (token.Type != Grammar.ERROR_TOKEN_TYPE)
				{
					Statistics.RecoveryTimes += 1;
					Statistics.RecoveryTimeSpent += DateTime.UtcNow - recoveryStartTime;

					return token;
				}
			}

			PotentialErrorMessage.Type = MessageType.Error;
			return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME, Grammar.ERROR_TOKEN_TYPE);
		}

		private bool StartsWithAny(Node subtree)
		{
			while (subtree.Symbol != Grammar.ANY_TOKEN_NAME
				&& subtree.Children.Count > 0)
				subtree = subtree.Children[0];

			return subtree.Symbol == Grammar.ANY_TOKEN_NAME;
		}

		private bool IsUnsafeAny(HashSet<string> oldStopTokens, string avoidedToken)
		{
			if (oldStopTokens != null && LexingStream.GetPairsCount() == NestingStack.Peek())
			{
				var anyArgs = Table.Items[StatesStack.Peek()].Markers
					.Where(i => i.Position == 0 && i.Next == Grammar.ANY_TOKEN_NAME)
					.Select(i => i.Alternative[0].Arguments)
					.FirstOrDefault();

				/*var nextState = Table[Stack.PeekState(), Grammar.ANY_TOKEN_NAME]
					.OfType<ShiftAction>().FirstOrDefault()
					.TargetItemIndex;*/

				Action shift = Table[StatesStack.Peek(), Grammar.ANY_TOKEN_NAME];

				return anyArgs.Contains(AnyArgument.Avoid, LexingStream.CurrentToken.Name)
					|| GetStopTokens(anyArgs, shift.TargetItemIndex).Except(oldStopTokens).Count() == 0
					&& (avoidedToken == null || anyArgs.Contains(AnyArgument.Avoid, avoidedToken));
			}

			return false;
		}
	}
}
