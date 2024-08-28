using Land.Core.Parsing.Tree;
using System;
using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows.Controls;

namespace Land.GUI.Serializers
{
	internal class GoSerializerCalls
	{
		internal static int ParseNode(Node root)
		{
			if (root == null)
				return 0;

			var str = root.ToString();
			if (str == "anon_func_call" || str=="call")
			{
				return 1;
			}
			var sum = 0;
			foreach (var cblock in root.Children)
			{
				sum += ParseNode(cblock);
			}
			return sum;
		}
		internal static void Serialize(string path, Node root)
		{
			FileInfo file = new FileInfo(path);
			file.Directory.Create();

			using (StreamWriter sw = File.CreateText(path))
			{
				var n = ParseNode(root);
				sw.WriteLine(JsonSerializer.Serialize(n));
			}

		}
	}
}
