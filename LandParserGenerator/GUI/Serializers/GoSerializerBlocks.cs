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
	internal class GoSerializerBlocks
	{
		internal static void ParseNode(Node root, GoBlock block)
		{
			if (root == null)
				return;

			var str = root.ToString();
			if (str == "water_entity")
			{
				return;
			}

			var blockCopy = block;
			if (str == "block")
			{
				var c = new GoBlock(block.Depth+1);
				block.Children.Add(c);
				blockCopy = c;
			}
			
			foreach (var subchild in root.Children)
			{
				ParseNode(subchild, blockCopy);
			}
		}
		internal static void Serialize(string path, Node root)
		{
			FileInfo file = new FileInfo(path);
			file.Directory.Create();

			using (StreamWriter sw = File.CreateText(path))
			{
				var n = new GoBlock(0);
				ParseNode(root, n);
				sw.WriteLine(JsonSerializer.Serialize(n));
			}

		}
	}
}
