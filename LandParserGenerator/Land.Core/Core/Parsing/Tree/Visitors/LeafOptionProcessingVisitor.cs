using System;
using System.Collections.Generic;
using System.Linq;
using Land.Core.Specification;
using Land.Core.Lexing;

namespace Land.Core.Parsing.Tree
{
	public class LeafOptionProcessingVisitor : GrammarProvidedTreeVisitor
	{
		public LeafOptionProcessingVisitor(Grammar g) : base(g) { }

		public override void Visit(Node node)
		{
			/// Если текущий узел должен быть листовым
			if (node.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LEAF) ||
				!node.Options.IsSet(NodeOption.GROUP_NAME) &&//.GetOptions(NodeOption.GROUP_NAME).Any() && 
				(GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LEAF, node.Symbol) ||
				!String.IsNullOrEmpty(node.Alias) &&
				GrammarObject.Options.IsSet(NodeOption.GROUP_NAME, NodeOption.LEAF, node.Alias)))
			{
				node.Value = FlattenValue(node);

				/// Перед тем, как удалить дочерние узлы,  вычисляем соответствие нового листа тексту
				var tmp = node.Location;

				node.Children.Clear();

				if (node.Location != null)
					node.SetLocation(tmp.Start, tmp.End);
			}
			else
				base.Visit(node);
		}

		private static List<string> FlattenValue(Node root)
		{
			// если у узла уже есть явное значение — просто скопируем его
			if (root.HasExplicitValue)
			{
				string txt;
				if (root.TryGetValueText(out txt))
					return String.IsNullOrEmpty(txt) ? new List<string>(0) : new List<string>(1) { txt };

				var v = root.Value;
				return v.Count == 0 ? new List<string>(0) : new List<string>(v);
			}

			var result = new List<string>(64);
			var stack = new Stack<Node>();
			stack.Push(root);

			while (stack.Count > 0)
			{
				var n = stack.Pop();

				if (n.HasExplicitValue)
				{
					string txt;
					if (n.TryGetValueText(out txt))
					{
						if (!String.IsNullOrEmpty(txt))
							result.Add(txt);
						continue;
					}

					var v = n.Value;
					if (v.Count > 0)
						result.AddRange(v);
					continue;
				}

				var ch = n.Children;
				// чтобы сохранить порядок слева-направо
				for (int i = ch.Count - 1; i >= 0; --i)
					stack.Push(ch[i]);
			}

			return result;
		}

	}
}
