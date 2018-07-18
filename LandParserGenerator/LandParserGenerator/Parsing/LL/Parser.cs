﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using LandParserGenerator.Lexing;
using LandParserGenerator.Parsing.Tree;

namespace LandParserGenerator.Parsing.LL
{
	public class Parser: BaseParser
	{
		private const int MAX_RECOVERY_ATTEMPTS = 5;

		private TableLL1 Table { get; set; }
		private Stack<Node> Stack { get; set; }
		private TokenStream LexingStream { get; set; }

		/// <summary>
		/// Стек открытых на момент прочтения последнего токена пар
		/// </summary>
		private Stack<PairSymbol> Nesting { get; set; }
		/// <summary>
		/// Уровень вложенности относительно описанных в грамматике пар,
		/// на котором начался разбор нетерминала
		/// </summary>
		private Dictionary<Node, int> NestingLevel { get; set; }

		public Parser(Grammar g, ILexer lexer): base(g, lexer)
		{
			Table = new TableLL1(g);

            /// В ходе парсинга потребуется First,
            /// учитывающее возможную пустоту ANY
            g.UseModifiedFirst = true;
		}

		/// <summary>
		/// LL(1) разбор
		/// </summary>
		/// <returns>
		/// Корень дерева разбора
		/// </returns>
		public override Node Parse(string text)
		{
			/// Логирование, статистика
			Log = new List<Message>();
			Statistics = new Statistics();
			var parsingStarted = DateTime.Now;

			/// Контроль вложенностей пар
			Nesting = new Stack<PairSymbol>();
			NestingLevel = new Dictionary<Node, int>();

            /// Готовим лексер и стеки
            LexingStream = new TokenStream(Lexer, text);
			Stack = new Stack<Node>();

			/// Кладём на стек стартовый символ
			var root = new Node(GrammarObject.StartSymbol);
			Stack.Push(new Node(Grammar.EOF_TOKEN_NAME));
			Stack.Push(root);

			/// Читаем первую лексему из входного потока
			var token = GetNextToken();

			/// Пока не прошли полностью правило для стартового символа
			while (Stack.Count > 0)
			{
				var stackTop = Stack.Peek();

				Log.Add(Message.Trace(
					$"Текущий токен: {GetTokenInfoForMessage(token)} | Символ на вершине стека: {GrammarObject.Userify(stackTop.Symbol)}",
					LexingStream.CurrentToken.Line, 
					LexingStream.CurrentToken.Column
				));

                /// Если символ на вершине стека совпадает с текущим токеном
                if(stackTop.Symbol == token.Name)
                {
					/// Снимаем узел со стека и устанавливаем координаты в координаты токена
					var node = Stack.Pop();

					/// Если текущий токен - признак пропуска символов, запускаем алгоритм
					if (token.Name == Grammar.ANY_TOKEN_NAME)
					{
						token = SkipAny(node, true);

						/// Если при пропуске текста произошла ошибка, прерываем разбор
						if (token.Name == Grammar.ERROR_TOKEN_NAME)
							break;
					}
					/// иначе читаем следующий токен
					else
					{
						node.SetAnchor(token.StartOffset, token.EndOffset);
						node.SetValue(token.Text);

						token = GetNextToken();
					}

					continue;
				}

				/// Если на вершине стека нетерминал, выбираем альтернативу по таблице
				if (GrammarObject[stackTop.Symbol] is NonterminalSymbol)
				{
					var alternatives = Table[stackTop.Symbol, token.Name];
					Alternative alternativeToApply = null;

					/// Сообщаем об ошибке в случае неоднозначной грамматики
					if (alternatives.Count > 1)
					{
						Log.Add(Message.Error(
							$"Неоднозначная грамматика: для нетерминала {GrammarObject.Userify(stackTop.Symbol)} и входного символа {GrammarObject.Userify(token.Name)} допустимо несколько альтернатив",
							token.Line,
							token.Column
						));
						break;
					}
					/// Если же в ячейке ровно одна альтернатива
					else if (alternatives.Count == 1)
					{
						alternativeToApply = alternatives.Single();
						Stack.Pop();

						if (!String.IsNullOrEmpty(alternativeToApply.Alias))
							stackTop.Alias = alternativeToApply.Alias;

						NestingLevel[stackTop] = Nesting.Count;

						for (var i = alternativeToApply.Count - 1; i >= 0; --i)
						{
							var newNode = new Node(alternativeToApply[i].Symbol, alternativeToApply[i].Options);

							stackTop.AddFirstChild(newNode);
							Stack.Push(newNode);
						}

						continue;
					}
				}

				/// Если не смогли ни сопоставить текущий токен с терминалом на вершине стека,
				/// ни найти ветку правила для нетерминала на вершине стека
				if (token.Name == Grammar.ANY_TOKEN_NAME)
				{
					Log.Add(Message.Warning(
						GrammarObject.Tokens.ContainsKey(stackTop.Symbol) ?
							$"Неожиданный символ {GetTokenInfoForMessage(LexingStream.CurrentToken)}, ожидался символ {GrammarObject.Userify(stackTop.Symbol)}" :
							$"Неожиданный символ {GetTokenInfoForMessage(LexingStream.CurrentToken)}, ожидался один из следующих символов: {String.Join(", ", Table[stackTop.Symbol].Where(t => t.Value.Count > 0).Select(t => GrammarObject.Userify(t.Key)))}",
						LexingStream.CurrentToken.Line,
						LexingStream.CurrentToken.Column
					));

					token = ErrorRecovery();

					if (token.Name == Grammar.ERROR_TOKEN_NAME)
						break;
				}
				/// Если непонятно, что делать с текущим токеном, и он конкретный
				/// (не Any), заменяем его на Any
				else
				{
					/// Если встретился неожиданный токен, но он в списке пропускаемых
					if (GrammarObject.Options.IsSet(ParsingOption.SKIP, token.Name))
					{
						token = GetNextToken();
					}
					else
					{
						token = Lexer.CreateToken(Grammar.ANY_TOKEN_NAME);
					}
				}
			}

			TreePostProcessing(root);

			Statistics.TimeSpent = DateTime.Now - parsingStarted;

			return root;
		}

		private IToken GetNextToken()
		{
			if (LexingStream.CurrentToken != null)
			{
				var token = LexingStream.CurrentToken;
				var closed = GrammarObject.Pairs.FirstOrDefault(p => p.Value.Right.Contains(token.Name));

				if (closed.Value != null && Nesting.Peek() == closed.Value)
					Nesting.Pop();

				var opened = GrammarObject.Pairs.FirstOrDefault(p => p.Value.Left.Contains(token.Name));

				if (opened.Value != null)
					Nesting.Push(opened.Value);
			}

			return LexingStream.NextToken();
		}

		private IToken GetNextToken(int level, out List<IToken> skipped)
		{
			skipped = new List<IToken>();

			while(true)
			{
				var next = GetNextToken();

				if (Nesting.Count == level || next.Name == Grammar.EOF_TOKEN_NAME)
					return next;
				else
					skipped.Add(next);
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
			var nestingCopy = new Stack<PairSymbol>(Nesting);
			var tokenIndex = LexingStream.CurrentIndex;

			IToken token = LexingStream.CurrentToken;
			HashSet<string> tokensAfterText;

			/// Если с Any не связана последовательность стоп-символов
			if (!anyNode.Options.AnyOptions.ContainsKey(AnyOption.Except))
			{
				/// Создаём последовательность символов, идущих в стеке после Any
				var alt = new Alternative();
				foreach (var elem in Stack)
					alt.Add(elem.Symbol);

				/// Определяем множество токенов, которые могут идти после Any
				tokensAfterText = GrammarObject.First(alt);
				/// Само Any во входном потоке нам и так не встретится, а вывод сообщения об ошибке будет красивее
				tokensAfterText.Remove(Grammar.ANY_TOKEN_NAME);

				/// Если указаны токены, которые нужно однозначно включать в Any
				if (anyNode.Options.AnyOptions.ContainsKey(AnyOption.Include))
				{
					tokensAfterText.ExceptWith(anyNode.Options.AnyOptions[AnyOption.Include]);
				}
			}
			else
			{
				tokensAfterText = anyNode.Options.AnyOptions[AnyOption.Except];
			}		

			/// Если Any непустой (текущий токен - это не токен,
			/// который может идти после Any)
			if (!tokensAfterText.Contains(token.Name))
			{
				/// Проверка на случай, если допропускаем текст в процессе восстановления
				if (!anyNode.StartOffset.HasValue)
					anyNode.SetAnchor(token.StartOffset, token.EndOffset);

				/// Смещение для участка, подобранного как текст
				int endOffset = token.EndOffset;
				var anyLevel = Nesting.Count;

				while (!tokensAfterText.Contains(token.Name)
					&& !anyNode.Options.Contains(AnyOption.Avoid, token.Name)
					&& token.Name != Grammar.EOF_TOKEN_NAME)
				{
					anyNode.Value.Add(token.Text);
					endOffset = token.EndOffset;

					if (anyNode.Options.AnyOptions.ContainsKey(AnyOption.IgnorePairs))
					{
						token = GetNextToken();
					}
					else
					{
						List<IToken> skippedBuffer;
						token = GetNextToken(anyLevel, out skippedBuffer);

						/// Если при пропуске до токена на том же уровне
						/// пропустили токены с более глубокой вложенностью
						if (skippedBuffer.Count > 0)
						{
							anyNode.Value.AddRange(skippedBuffer.Select(t => t.Text));
							endOffset = skippedBuffer.Last().EndOffset;
						}
					}
				}

				anyNode.SetAnchor(anyNode.StartOffset.Value, endOffset);

				/// Если дошли до конца входной строки, и это было не по плану
				if (token.Name == Grammar.EOF_TOKEN_NAME && !tokensAfterText.Contains(token.Name)
					|| anyNode.Options.Contains(AnyOption.Avoid, token.Name))
				{
					var message = Message.Trace(
						$"Ошибка при пропуске {Grammar.ANY_TOKEN_NAME}: неожиданный токен {GrammarObject.Userify(token.Name)}, ожидался один из следующих символов: { String.Join(", ", tokensAfterText.Select(t => GrammarObject.Userify(t))) }",
						token.Line,
						token.Column
					);

					if (enableRecovery)
					{
						message.Type = MessageType.Warning;
						Log.Add(message);

						Nesting = nestingCopy;
						LexingStream.MoveTo(tokenIndex);
						anyNode.Reset();
						///	Возвращаем узел обратно на стек
						Stack.Push(anyNode);

						return ErrorRecovery();
					}
					else
					{
						message.Type = MessageType.Error;
						Log.Add(message);
						return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME);
					}
				}
			}

			return token;
		}

		private IToken ErrorRecovery()
		{
			Log.Add(Message.Warning(
				$"Процесс восстановления запущен в позиции токена {GetTokenInfoForMessage(LexingStream.CurrentToken)}",
				LexingStream.CurrentToken.Line,
				LexingStream.CurrentToken.Column
			));

			/// То, что мы хотели разобрать, и не смогли
			var currentNode = Stack.Peek();
			Stack.Pop();

			/// Если произошла ошибка при пропуске Any и конфигурация ветки
			/// может заставить нас повторно перейти на неё же при восстановлении,
			/// переходим сразу к родителю родителя этого Any
			if(currentNode.Symbol == Grammar.ANY_TOKEN_NAME 
				&& currentNode.Parent.Children.Count == 1)
			{
				currentNode = currentNode.Parent.Parent;
			}

			/// Поднимаемся по уже построенной части дерева, пока не встретим узел нетерминала,
			/// для которого допустима альтернатива из одного Any
			while (currentNode != null
				&& (!GrammarObject.Rules.ContainsKey(currentNode.Symbol)
				|| !GrammarObject.Rules[currentNode.Symbol].Alternatives.Any(a => a.Count == 1 && a[0].Symbol == Grammar.ANY_TOKEN_NAME)))
			{
				if (currentNode.Parent != null)
				{
					var childIndex = currentNode.Parent.Children.IndexOf(currentNode);

					/// Снимаем со стека всех неразобранных потомков родителя текущего узла,
					/// для текущего узла они являются правыми братьями
					for (var i = 0; i < currentNode.Parent.Children.Count - childIndex - 1; ++i)
						Stack.Pop();
				}

				/// Переходим к родителю
				currentNode = currentNode.Parent;
			}

			if(currentNode != null)
			{
				List<IToken> skippedBuffer;

				if (Nesting.Count != NestingLevel[currentNode])
				{
					/// Запоминаем токен, на котором произошла ошибка
					var errorToken = LexingStream.CurrentToken;
					/// Пропускаем токены, пока не поднимемся на тот же уровень вложенности,
					/// на котором раскрывали нетерминал
					GetNextToken(NestingLevel[currentNode], out skippedBuffer);
					skippedBuffer.Insert(0, errorToken);
				}
				else
				{
					skippedBuffer = new List<IToken>();
				}

				var alternativeToApply = Table[currentNode.Symbol, Grammar.ANY_TOKEN_NAME][0];
				var anyNode = new Node(alternativeToApply[0].Symbol, alternativeToApply[0].Options);

				anyNode.Value = currentNode.GetValue();
				anyNode.Value.AddRange(skippedBuffer.Select(t => t.Text));

				if (currentNode.StartOffset.HasValue)
					anyNode.SetAnchor(currentNode.StartOffset.Value, currentNode.EndOffset.Value);
				if (skippedBuffer.Count > 0)
				{
					anyNode.SetAnchor(
						anyNode.StartOffset.HasValue ? anyNode.StartOffset.Value : skippedBuffer.First().StartOffset,
						skippedBuffer.Last().EndOffset
					);
				}

				/// Пытаемся пропустить Any в этом месте
				var token = SkipAny(anyNode, false);

				/// Если Any успешно пропустили и возобновили разбор,
				/// возвращаем токен, с которого разбор продолжается
				if (token.Name != Grammar.ERROR_TOKEN_NAME)
				{
					currentNode.ResetChildren();
					currentNode.AddFirstChild(anyNode);

					if (!String.IsNullOrEmpty(alternativeToApply.Alias))
						currentNode.Alias = alternativeToApply.Alias;

					Log.Add(Message.Warning(
						$"Произведено восстановление на уровне {currentNode.Symbol}, разбор продолжен с токена {GetTokenInfoForMessage(token)}",
						token.Line,
						token.Column
					));

					return token;
				}
			}

			Log.Add(Message.Error(
				$"Не удалось продолжить разбор",
				null
			));

			return Lexer.CreateToken(Grammar.ERROR_TOKEN_NAME);
		}
	}
}
