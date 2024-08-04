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

			var str = root.ToString();
			if (str == "content")
			{
				node.Type = "func_body";
			}

			if (str == "call")
			{
				var c = new GoNode(str, root.Children[0].ToString().Replace("ID: ", ""));
				node.Children.Add(c);
				foreach (var child in root.Children)
				{
					ParseNode(child, c);
				}
				return;
			}
			if (str == "if" || str == "switch" || str == "select")
			{
				var c = new GoNode(str);
				node.Children.Add(c);
				foreach (var child in root.Children)
				{
					ParseNode(child, c);
				}
				return;
			}
			foreach (var cblock in root.Children)
			{
				ParseNode(cblock, node);
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
