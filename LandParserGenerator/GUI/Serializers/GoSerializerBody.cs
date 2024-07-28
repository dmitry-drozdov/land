using Land.Core.Parsing.Tree;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows.Controls;

namespace Land.GUI.Serializers
{
	internal class GoSerializerBody
	{
		internal static void ParseNode(Node root, GoNode node)
		{
			if (root == null || root.ToString() == "water_entity")
				return;
			if (root.ToString() == "content")
			{
				node.Type = "func_body";
			}

			foreach (var child in root.Children)
			{
				var str = child.ToString();
				if (str == "call" || str == "if" || str == "switch" || str == "select")
				{
					var c = new GoNode(str);
					node.Children.Add(c);
					ParseNode(child, c);
				}
				if (str == "block" || str == "control")
				{
					foreach (var cblock in child.Children)
					{
						ParseNode(cblock, node);
					}
				}
			}

		}
		internal static void Serialize(string path, Node root)
		{
			FileInfo file = new FileInfo(path);
			file.Directory.Create();

			using (StreamWriter sw = File.CreateText(path))
			{
				var n = new GoNode("");
				ParseNode(root, n);
				sw.WriteLine(JsonSerializer.Serialize(n));
			}

		}
	}
}
