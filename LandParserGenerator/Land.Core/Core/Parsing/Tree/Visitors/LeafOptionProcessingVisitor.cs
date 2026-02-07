using System;
using System.Collections.Generic;
using Land.Core.Specification;
using Land.Core.Lexing;

namespace Land.Core.Parsing.Tree
{
	public class LeafOptionProcessingVisitor : GrammarProvidedTreeVisitor
	{
		// Reused buffers to avoid per-leaf allocations (hot path).
		// Visitor is executed sequentially during post-processing.
		private readonly Stack<Node> _stack = new Stack<Node>(64);
		private readonly List<string> _buffer = new List<string>(64);

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
				// Before dropping children, ensure location is computed.
				// (Location may be lazy and depends on children.)
				var _ = node.Location;

				FlattenInto(node);

				// Drop children in O(1) (Clear() is O(n) and shows up in profiles for large lists).
				node.Children = null;
			}
			else
				base.Visit(node);
		}

		/// <summary>
		/// Flattens all explicit values under <paramref name="root"/> into <paramref name="root"/>
		/// while avoiding allocations for the common case (single terminal value).
		/// </summary>
		private void FlattenInto(Node root)
		{
			// If node already has explicit value, keep it as-is.
			// Previous implementation copied it into a new List<string>, losing the single-value fast-path.
			if (root.HasExplicitValue)
				return;

			_buffer.Clear();
			_stack.Clear();
			_stack.Push(root);

			string first = null;
			bool hasFirst = false;
			bool usingBuffer = false;

			while (_stack.Count > 0)
			{
				var n = _stack.Pop();

				// Handle lazy Any token range without materializing List<string>.
				TokenStream stream;
				int startIndex, endExclusive;
				if (n.TryGetLazyTokenRange(out stream, out startIndex, out endExclusive))
				{
					for (int i = startIndex; i < endExclusive; i++)
						AddValueInline(stream.GetTokenAt(i).Text, ref first, ref hasFirst, ref usingBuffer);
					continue;
				}

				if (n.HasExplicitValue)
				{
					string txt;
					if (n.TryGetValueText(out txt))
					{
						AddValueInline(txt, ref first, ref hasFirst, ref usingBuffer);
						continue;
					}

					var v = n.Value;
					if (v.Count > 0)
					{
						for (int i = 0; i < v.Count; i++)
							AddValueInline(v[i], ref first, ref hasFirst, ref usingBuffer);
					}
					continue;
				}

				var ch = n.Children;
				// Preserve left-to-right order.
				for (int i = ch.Count - 1; i >= 0; --i)
					_stack.Push(ch[i]);
			}

			if (usingBuffer)
			{
				// Copy buffer into a node-owned list.
				root.Value = new List<string>(_buffer);
			}
			else if (hasFirst)
			{
				root.SetValueText(first);
			}
			else
			{
				// Explicit empty value (matches previous semantics).
				root.SetValue();
			}
		}

		private void AddValueInline(string s, ref string first, ref bool hasFirst, ref bool usingBuffer)
		{
			if (String.IsNullOrEmpty(s))
				return;

			if (!usingBuffer)
			{
				if (!hasFirst)
				{
					first = s;
					hasFirst = true;
					return;
				}

				usingBuffer = true;
				_buffer.Clear();
				_buffer.Add(first);
				_buffer.Add(s);
				first = null;
				hasFirst = false;
				return;
			}

			_buffer.Add(s);
		}

	}
}
