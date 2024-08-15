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
	internal class GoSerializerBrakcets
	{
		internal static void ParseNode(Node root, GoNode node)
		{
			if (root == null)
				return;

			var str = root.ToString();
			if (str == "water_entity")
			{
				var c = new GoNode("any");
				node.Children.Add(c);
				return;
			}
			if (str == "content")
			{
				node.Type = "root";
			}
		

			if (str == "block")
			{
				var c = new GoNode(str);
				node.Children.Add(c);
				foreach (var child in root.Children)
				{
					ParseNode(child, c);
				}
				return;
			}
			if (str == "if")
			{
				var c = new GoNode(str);
				node.Children.Add(c);
				foreach (var child in root.Children)
				{
					if (child.ToString() == "block") // проваливаемся внутрь block, это д.б. часть if-а
					{
						foreach (var subchild in child.Children)
						{
							ParseNode(subchild, c);
						}
					}
					else
					{
						ParseNode(child, c);
					}
				}
				return;
			}

			foreach (var subchild in root.Children)
			{
				ParseNode(subchild, node);
			}
			return;
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
