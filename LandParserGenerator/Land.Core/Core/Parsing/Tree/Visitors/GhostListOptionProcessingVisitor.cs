using System;
using System.Collections.Generic;
using Land.Core.Specification;

namespace Land.Core.Parsing.Tree
{
	/// <summary>
	/// Обработка опций:
	/// 1) VOID  -> удалить подузел
	/// 2) GHOST -> поднять детей подузла на уровень выше (вставка на место)
	/// 3) LIST  -> "сплющить" подузлы того же типа (Symbol/Alias)
	///
	/// Оптимизация:
	/// - без RemoveAt/Insert в середине List (избавляемся от O(n^2))
	/// - без GetOptions(...).Any() (избавляемся от аллокаций List<string>)
	/// - одно прохождение + worklist-раскрытие, чтобы применить правила транзитивно
	///   к вставляемым узлам (как в исходной реализации с --i)
	/// </summary>
	public class GhostListOptionProcessingVisitor : GrammarProvidedTreeVisitor
	{
		public GhostListOptionProcessingVisitor(Grammar g) : base(g) { }

		public override void Visit(Node node)
		{
			var children = node.Children;
			if (children.Count == 0)
			{
				base.Visit(node);
				return;
			}

			// Глобальные LIST-опции для символа/алиаса текущего узла
			var listForAlias = !String.IsNullOrEmpty(node.Alias)
				&& GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LIST, node.Alias);
			var listForSymbol = GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LIST, node.Symbol);

			// Локальная LIST-опция (если в узле есть локальная группа опций)
			var nodeHasLocalGroup = node.Options.IsSet(NodeOption.GROUP_NAME);
			var nodeIsList =
				node.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LIST)
				|| (!nodeHasLocalGroup && (listForAlias || listForSymbol));

			// Если ни один ребёнок не подпадает под правила, не перестраиваем список
			List<Node> newChildren = null;

			for (int i = 0; i < children.Count; i++)
			{
				var child = children[i];

				if (newChildren == null)
				{
					if (!NeedsRewrite(node, child, nodeIsList, listForSymbol, listForAlias))
						continue;

					// Первый случай изменения: создаём новый список и копируем то, что уже прошло без изменений
					newChildren = new List<Node>(children.Count);
					for (int k = 0; k < i; k++)
					{
						var c = children[k];
						c.Parent = node;
						newChildren.Add(c);
					}
				}

				// Если изменения уже начались, переписываем/раскрываем текущего ребёнка
				if (newChildren != null)
				{
					Emit(node, child, nodeIsList, listForSymbol, listForAlias, newChildren);
				}
			}

			// Если были изменения — заменяем детей
			if (newChildren != null)
				node.Children = newChildren;

			base.Visit(node);
		}

		private bool NeedsRewrite(Node parent, Node child, bool parentIsList, bool listForSymbol, bool listForAlias)
		{
			// VOID / GHOST
			if (IsOptionSet(NodeOption.VOID, child)) return true;
			if (IsOptionSet(NodeOption.GHOST, child)) return true;

			// LIST flatten
			if (parentIsList)
			{
				if (listForSymbol && child.Symbol == parent.Symbol) return true;
				if (listForAlias && child.Alias == parent.Alias) return true;
			}

			return false;
		}

		/// <summary>
		/// Эмитит в output один или несколько узлов после применения правил VOID/GHOST/LIST.
		/// Раскрывает правила транзитивно (как исходный цикл с Insert + --i).
		/// </summary>
		private void Emit(Node parent, Node start, bool parentIsList, bool listForSymbol, bool listForAlias, List<Node> output)
		{
			// Worklist: стек обеспечивает сохранение порядка при пуше детей в обратном порядке
			var stack = new Stack<Node>();
			stack.Push(start);

			while (stack.Count > 0)
			{
				var node = stack.Pop();

				// 1) VOID: удалить
				if (IsOptionSet(NodeOption.VOID, node))
					continue;

				// 2) GHOST: поднять детей
				if (IsOptionSet(NodeOption.GHOST, node))
				{
					var ch = node.Children;
					for (int j = ch.Count - 1; j >= 0; j--)
						stack.Push(ch[j]);
					continue;
				}

				// 3) LIST flatten: убрать подузел того же типа
				if (parentIsList &&
					((listForSymbol && node.Symbol == parent.Symbol) ||
					 (listForAlias && node.Alias == parent.Alias)))
				{
					var ch = node.Children;
					for (int j = ch.Count - 1; j >= 0; j--)
						stack.Push(ch[j]);
					continue;
				}

				node.Parent = parent;
				output.Add(node);
			}
		}

		/// <summary>
		/// Опция считается установленной, если:
		/// - локально на узле стоит group/option, ИЛИ
		/// - на узле нет локальных опций в группе, но глобально опция задана для symbol/alias
		///
		/// Оптимизация: не используем GetOptions(...).Any() (он аллоцирует список).
		/// Вместо этого — IsSet(group).
		/// </summary>
		private bool IsOptionSet(string option, Node nodeToCheck)
		{
			// локально
			if (nodeToCheck.Options.IsSet(NodeOption.GROUP_NAME, option))
				return true;

			// если локально задана группа опций, то глобальные не применяются
			if (nodeToCheck.Options.IsSet(NodeOption.GROUP_NAME))
				return false;

			// глобально по символу/алиасу
			if (GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, option, nodeToCheck.Symbol))
				return true;

			return !String.IsNullOrEmpty(nodeToCheck.Alias)
				&& GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, option, nodeToCheck.Alias);
		}
	}
}
