﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Land.Core.Specification;
using Land.Core.Lexing;
using Land.Core.Parsing.Tree;
using System.Diagnostics;

namespace Land.Core.Parsing.LR
{
	public class Parser : BaseParser
	{
		private TableLR1 Table { get; set; }

		private ParsingStack Stack { get; set; }
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

			var d = new Durations();

			var watcher  = Stopwatch.StartNew();

			d.Start();

			/// Множество индексов токенов, на которых запускалось восстановление
			PositionsWhereRecoveryStarted = new HashSet<int>();
			/// Создаём стек для уровней вложенности пар
			NestingStack = new Stack<int>();
			/// Готовим лексер
			LexingStream = new ComplexTokenStream(GrammarObject, Lexer, text, Log);
			/// Читаем первую лексему из входного потока
			var token = LexingStream.GetNextToken();
			/// Создаём стек
			Stack = new ParsingStack();
			Stack.Push(0);
			NestingStack.Push(0);

			d.Stop("init");


			while (true)
			{
				if (token.Type == Grammar.ERROR_TOKEN_TYPE)
					break;

				d.Start();
				var currentState = Stack.PeekState();
				d.Stop("PeekState");

				if (EnableTracing && token.Type != Grammar.ERROR_TOKEN_TYPE && token.Type != Grammar.ANY_TOKEN_TYPE)
					Log.Add(Message.Trace(
						$"Текущий токен: {this.Developerify(token)} | Стек: {Stack.ToString(GrammarObject)}",
						token.Location.Start
					));

				if (Table[currentState, token.Name].Count > 0)
				{
					if (token.Type == Grammar.ANY_TOKEN_TYPE)
					{
						d.Start();
						token = SkipAny(NodeGenerator.Generate(Grammar.ANY_TOKEN_NAME), null, true);
						d.Stop("SkipAny");

						/// Если при пропуске текста произошла ошибка, прерываем разбор
						if (token.Type == Grammar.ERROR_TOKEN_TYPE)
							break;
						else
							continue;
					}

					d.Start();
					var action = GetAction(currentState, token.Name);
					d.Stop("GetAction");

					d.Start();
					/// Если нужно произвести перенос
					if (action is ShiftAction)
					{
						var tokenNode = NodeGenerator.Generate(token.Name);
						tokenNode.SetValue(token.Text);
						tokenNode.SetLocation(token.Location.Start, token.Location.End);

						var shift = (ShiftAction)action;
						/// Вносим в стек новое состояние
						Stack.Push(tokenNode, shift.TargetItemIndex);
						NestingStack.Push(LexingStream.GetPairsCount());

						if (EnableTracing)
						{
							Log.Add(Message.Trace(
								$"Перенос",
								token.Location.Start
							));
						}

						token = LexingStream.GetNextToken();
						d.Stop("ShiftAction");
					}
					/// Если нужно произвести свёртку
					else if (action is ReduceAction reduce)
					{
						var parentNode = NodeGenerator.Generate(reduce.ReductionAlternative.NonterminalSymbolName);

						/// Снимаем со стека символы ветки, по которой нужно произвести свёртку
						for (var i = 0; i < reduce.ReductionAlternative.Count; ++i)
						{
							parentNode.AddFirstChild(Stack.PeekSymbol());
							Stack.Pop();
							NestingStack.Pop();
						}
						currentState = Stack.PeekState();

						/// Кладём на стек состояние, в которое нужно произвести переход
						Stack.Push(
							parentNode,
							Table.Transitions[currentState][reduce.ReductionAlternative.NonterminalSymbolName]
						);
						NestingStack.Push(LexingStream.GetPairsCount());

						if (EnableTracing)
						{
							Log.Add(Message.Trace(
								$"Свёртка по правилу {GrammarObject.Developerify(reduce.ReductionAlternative)} -> {GrammarObject.Developerify(reduce.ReductionAlternative.NonterminalSymbolName)}",
								token.Location.Start
							));
						}
						d.Stop("ReduceAction");
						continue;
					}
					else if (action is AcceptAction)
					{
						root = Stack.PeekSymbol();
						d.Stop("PeekSymbol");
						break;
					}
				}
				else if (token.Type == Grammar.ANY_TOKEN_TYPE)
				{
					d.Start();

					Log.Add(PotentialErrorMessage = Message.Trace(
						$"Неожиданный символ {this.Developerify(LexingStream.CurrentToken)} для состояния{Environment.NewLine}\t\t" + Table.ToString(Stack.PeekState(), null, "\t\t"),
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
								Table.Items[Stack.PeekState()].Markers
									.Where(i=>i.Lookahead != null).Select(e => e.Lookahead).ToList()
							}
						}
					));


					token = ErrorRecovery();
					d.Stop("ErrorRecovery");
				}
				else
				{
					d.Start();
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
					d.Stop("Unexpected token");
				}
			}

			if (root != null)
			{
				d.Start();
				root = TreePostProcessing(root);
				d.Stop("TreePostProcessing");
			}

			d.Add("ParsingAlgorithm", watcher);

			return (root, d);
		}

		private Action GetAction(int currentState, string token)
		{
			if (Table[currentState, token].Count == 0)
				return null;

			return Table[currentState, token].OfType<ShiftAction>().FirstOrDefault()
				?? Table[currentState, token].First();
		}

		private IToken SkipAny(Node anyNode, Durations d, bool enableRecovery)
		{
			var nestingCopy = LexingStream.GetPairsState();
			var token = LexingStream.CurrentToken;
			var tokenIndex = LexingStream.CurrentIndex;
			var rawActions = Table[Stack.PeekState(), Grammar.ANY_TOKEN_NAME];

			if (EnableTracing)
			{
				Log.Add(Message.Trace(
					$"Инициирован пропуск Any | Стек: {Stack.ToString(GrammarObject)} | Состояние: {Environment.NewLine}\t\t"
						+ Table.ToString(Stack.PeekState(), null, "\t\t"),
					token.Location.Start
				));
			}

			/// Пока по Any нужно производить свёртки (ячейка таблицы непуста и нет конфликтов)
			while (rawActions.Count == 1 && rawActions.First() is ReduceAction reduce)
			{
				var parentNode = NodeGenerator.Generate(reduce.ReductionAlternative.NonterminalSymbolName);

				/// Снимаем со стека символы ветки, по которой нужно произвести свёртку
				for (var i = 0; i < reduce.ReductionAlternative.Count; ++i)
				{
					parentNode.AddFirstChild(Stack.PeekSymbol());
					Stack.Pop();
					NestingStack.Pop();
				}

				/// Кладём на стек состояние, в которое нужно произвести переход
				Stack.Push(
					parentNode,
					Table.Transitions[Stack.PeekState()][reduce.ReductionAlternative.NonterminalSymbolName]
				);
				NestingStack.Push(LexingStream.GetPairsCount());

				rawActions = Table[Stack.PeekState(), Grammar.ANY_TOKEN_NAME];
			}

			/// Берём опции из нужного вхождения Any
			var marker = Table.Items[Stack.PeekState()].Markers
				.First(i => i.Next == Grammar.ANY_TOKEN_NAME);

			anyNode.Options = marker.Alternative[marker.Position].Options;
			anyNode.Arguments = marker.Alternative[marker.Position].Arguments;

			/// Проверяем, не происходит ли восстановление в действительно некорректной программе
			if (anyNode.Arguments.Contains(AnyArgument.Error)
				&& PotentialErrorMessage != null)
			{
				PotentialErrorMessage.Type = MessageType.Error;
			}

			/// Производим перенос
			var shift = (ShiftAction)rawActions.Where(a => a is ShiftAction).Single();
			/// Вносим в стек новое состояние
			Stack.Push(anyNode, shift.TargetItemIndex);
			NestingStack.Push(LexingStream.GetPairsCount());


			if (EnableTracing)
			{
				Log.Add(Message.Trace(
					$"Поиск окончания последовательности, соответствующей Any | Стек: {Stack.ToString(GrammarObject)} | Состояние: {Environment.NewLine}\t\t"
						+ Table.ToString(Stack.PeekState(), null, "\t\t"),
					token.Location.Start
				));
			}

			var stopTokens = GetStopTokens(anyNode.Arguments, Stack.PeekState());
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
					token = LexingStream.GetNextToken(anyLevel, d, out List<IToken> skippedBuffer);

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
			var stopTokens = args.Contains(AnyArgument.Except)
				? args.AnyArguments[AnyArgument.Except]
				: new HashSet<string>(
					Table.GetExpectedTokens(state).Except(args.Contains(AnyArgument.Include)
						? args.AnyArguments[AnyArgument.Include] : new HashSet<string>())
				);

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
				if (Stack.CountSymbols > 0)
				{
					if (Stack.PeekSymbol().Location != null)
					{
						startLocation = Stack.PeekSymbol().Location.Start;
						if (endLocation == null)
						{
							endLocation = Stack.PeekSymbol().Location.End;
						}
					}

					value = Stack.PeekSymbol().GetValue()
						.Concat(value).ToList();

					// Запоминаем снятый со стека символ - это то, что было успешно распознано
					previouslyMatched = Stack.PeekSymbol();
				}

				Stack.Pop();
				NestingStack.Pop();

				if (Stack.CountStates > 0)
				{
					// Выбираем пункты, продукции которых потенциально могут участвовать
					// в выводе текущего префикса из стартового символа
					initialDerivationProds = new HashSet<PathFragment>(
						Table.Items[Stack.PeekState()].Markers
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
						derivationProds.UnionWith(Table.Items[Stack.PeekState()].Markers
							.Where(i => derivationProds.Any(p => p.Pos == 0 && p.Alt.NonterminalSymbolName == i.Next))
							.Select(i => new PathFragment { Alt = i.Alternative, Pos = i.Position })
						);
					}
				}
			}
			while (Stack.CountStates > 0 && (derivationProds.Count == initialDerivationProds.Count
				|| derivationProds.Except(initialDerivationProds).All(p => !GrammarObject.Options.IsSet(ParsingOption.GROUP_NAME, ParsingOption.RECOVERY, p.Alt[p.Pos]))
				|| StartsWithAny(previouslyMatched)
				|| IsUnsafeAny(stopTokens, avoidedToken))
			);

			if (Stack.CountStates > 0)
			{
				if (LexingStream.GetPairsCount() != NestingStack.Peek())
				{
					var skippedBuffer = new List<IToken>();

					// Запоминаем токен, на котором произошла ошибка
					var currentToken = LexingStream.CurrentToken;

					// Пропускаем токены, пока не поднимемся на тот же уровень вложенности, 
					// на котором раскрывали нетерминал
					var nonterminalLevelToken = LexingStream.GetNextToken(NestingStack.Peek(), new Durations(), out skippedBuffer);

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

				Log.Add(Message.Trace(
					$"Попытка продолжить разбор в состоянии {Environment.NewLine}\t\t{Table.ToString(Stack.PeekState(), null, "\t\t")}\tв позиции токена {this.Developerify(LexingStream.CurrentToken)}",
					LexingStream.CurrentToken.Location.Start
				));

				var token = SkipAny(anyNode, new Durations(), false);

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
				var anyArgs = Table.Items[Stack.PeekState()].Markers
					.Where(i => i.Position == 0 && i.Next == Grammar.ANY_TOKEN_NAME)
					.Select(i => i.Alternative[0].Arguments)
					.FirstOrDefault();

				var nextState = Table[Stack.PeekState(), Grammar.ANY_TOKEN_NAME]
					.OfType<ShiftAction>().FirstOrDefault()
					.TargetItemIndex;

				return anyArgs.Contains(AnyArgument.Avoid, LexingStream.CurrentToken.Name)
					|| GetStopTokens(anyArgs, nextState).Except(oldStopTokens).Count() == 0
					&& (avoidedToken == null || anyArgs.Contains(AnyArgument.Avoid, avoidedToken));
			}

			return false;
		}
	}
}
