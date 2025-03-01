using Land.Core.Parsing.Tree;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Land.GUI.Visitor
{
	internal class TypeScriptVisitor
	{

		private static void Debug(string msg)
		{
			System.Diagnostics.Debug.WriteLine("LOG🔔 " + msg);
		}
		internal static void CheckAndDumpData(string path, Node root)
		{
			var d = new Dictionary<string, int> {
				{ "ResolverTest_class_1_func_1", 0 },
				{ "ResolverTest_class_1_func_2", 0 },
				{ "ResolverTest_class_2_func_1", 0 },
				{ "ResolverTest_lambda_1_func_1", 0 },
				{ "ResolverTest_lambda_1_func_2", 0 },
				{ "ResolverTest_export_1_func_1", 0 },
				{ "ResolverTest_export_1_func_2", 0 },
				{ "ResolverTest_struct_1_func_1", 0 },
				{ "ResolverTest_struct_1_func_2", 0 },
				{ "ResolverTest_struct_1_func_3", 0 },
				{ "ResolverTest_struct_1_func_4", 0 },
				{ "ResolverTest_struct_1_func_5", 0 },
				{ "ResolverTest_struct_1_func_6", 0 },
				{ "ResolverTest_struct_1_func_7", 0 },
				{ "ResolverTest_struct_1_func_8", 0 },
				{ "ResolverTest_struct_1_func_9", 0 },
				{ "ResolverTest_struct_1_func_10", 0 },
				{ "ResolverTest_struct_1_func_11", 0 },
				{ "ResolverTest_struct_1_func_12", 0 },
			};
			var l = new List<string>();
			Visit(root, d, l);
			bool ok = true;
			foreach (var item in d)
			{
				if (item.Value != 1)
				{
					Debug($"incorrect count {path}: {item.Key} {item.Value}");
					ok = false;
				}
			}
			if (ok)
			{
				//Debug($"OK {path}");
			}
			Dump(path, l);
		}

		internal static void Dump(string path, List<string> l)
		{
			FileInfo file = new FileInfo(path);
			file.Directory.Create();

			using (StreamWriter sw = File.CreateText(path))
			{
				sw.Write(string.Join("\n", l));
			}
		}

		internal static void Visit(Node root, Dictionary<string, int> d, List<string> l)
		{
			var nodeName = root.ToString();
			if (nodeName == "water_entity")
			{
				return;
			}
			if (nodeName == "func" || nodeName == "resolver_line_impl" || nodeName == "resolver_line_var")
			{
				var funcName = root.Children[0].ToString().Replace("ID: ", "");
				if (d.ContainsKey(funcName))
				{
					d[funcName]++;
				}
				l.Add(funcName);
				return;
			}
			foreach (var child in root.Children)
			{
				if (child.ToString() == "water_entity")
				{
					continue;
				}
				Visit(child, d, l);
			}
		}

	}
}
