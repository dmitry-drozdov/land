﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace LandParserGenerator.Parsing.Tree
{
	public class Node
	{
		public Node Parent { get; set; }

		public string Symbol { get; set; }
		public List<string> Value { get; set; } = new List<string>();
		public List<Node> Children { get; private set; } = new List<Node>();

		public NodeOption ProcessingOption { get; set; } = NodeOption.NONE;

		private Location Anchor { get; set; }
		private bool AnchorReady { get; set; }

		public int? StartOffset
		{
			get
			{
				if (Anchor == null && !AnchorReady)
					GetAnchorFromChildren();
				return Anchor?.StartOffset;
			}
		}
		public int? EndOffset
		{
			get
			{
				if (Anchor == null && !AnchorReady)
					GetAnchorFromChildren();
				return Anchor?.EndOffset;
			}
		}

		private void GetAnchorFromChildren()
		{
			if (Children.Count > 0)
			{
				Anchor = Children[0].Anchor;

				foreach (var child in Children)
				{
					if (child.Anchor == null)
						child.GetAnchorFromChildren();

					if (Anchor == null)
						Anchor = child.Anchor;
					else
						Anchor = Anchor.Merge(child.Anchor);
				}
			}

			AnchorReady = true;
		}

		public Node(string smb)
		{
			Symbol = smb;
		}

		public Node(string smb, NodeOption opt)
		{
			Symbol = smb;
			ProcessingOption = opt;
		}

		public void AddLastChild(Node child)
		{
			Children.Add(child);
			child.Parent = this;
		}

		public void AddFirstChild(Node child)
		{
			Children.Insert(0, child);
			child.Parent = this;
		}

		public void ResetChildren()
		{
			Children = new List<Node>();
		}

		public void SetAnchor(int start, int end)
		{
			AnchorReady = true;

			Anchor = new Location()
			{
				StartOffset = start,
				EndOffset = end
			};
		}

		public void SetValue(params string[] vals)
		{
			Value = new List<string>(vals);
		}

		public void Accept(BaseVisitor visitor)
		{
			visitor.Visit(this);
		}

		public override string ToString()
		{
			return this.Symbol;
		}
	}
}
