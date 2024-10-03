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
	internal class GoSerializerControl
	{
		internal static void ParseNode(Node root, GoControl node)
		{
			if (root == null || root.ToString() == "water_entity")
				return;

			var str = root.ToString();
			if (str == "content")
			{
				node.Type = "root";
			}

			if (str == "if")
			{
				var c = new GoControl(str, node.Depth + 1);
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
				var n = new GoControl("root", 0);
				ParseNode(root, n);

				JsonSerializerOptions options = new JsonSerializerOptions();

				options.ReferenceHandler = System.Text.Json.Serialization.ReferenceHandler.IgnoreCycles;
				options.WriteIndented = true;


				sw.WriteLine(JsonSerializer.Serialize(n, options));
			}

		}
	}
}
