using System;
using System.Collections.Generic;
using System.Linq;
using System.Xml.Linq;
using Land.Core.Specification;

namespace Land.Core.Parsing.Tree
{
	public class TreeManager
	{
		public static void MergeTrees(Node node1, Node node2)
		{
			var controls = AllControlNodes(node1).OrderBy(x => x.Location?.Start?.Offset ?? 0).ToList();
			var calls = AllCallsNodes(node2);

			foreach (var call in calls)
			{
				Node foundControl = null;
				foreach (var control in controls)
				{
					if (call.Location.HasOverlap(control.Location))
					{
						//throw new InvalidOperationException("overlapping detected");
						continue;
					}

					if (control.Location?.Includes(call.Location) == true)
					{
						foundControl = control;
					}
				}
				if (foundControl == null)
				{
					//throw new InvalidOperationException("cannot weave call in control");
					continue;
				}
				call.Parent = foundControl;
				foundControl.Children.Add(call);
				foundControl.Children = foundControl.Children.OrderBy(x => x.Location?.Start.Offset ?? 0).ToList();
			}
		}

		public static List<Node> AllControlNodes(Node root)
		{
			if (root == null) return null;
			var list = new List<Node> { root };
			foreach (var node in root.Children)
			{
				list.AddRange(AllControlNodes(node));
			}

			return list;
		}

		public static List<Node> AllCallsNodes(Node root)
		{
			var list = new List<Node>();
			if (root == null) return list;

			foreach (var node in root.Children)
			{
				var str = node.ToString();
				if (str == "call")
				{
					list.Add(node);
					continue;
				}
				list.AddRange(AllCallsNodes(node));
			}

			return list;
		}
	}
}
