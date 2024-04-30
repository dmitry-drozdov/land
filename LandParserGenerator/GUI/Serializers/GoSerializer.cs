using Land.Core.Parsing.Tree;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Land.GUI.Serializers
{
	internal class GoSerializer
	{
		static void ParseStruct(Node pc, StreamWriter sw)
		{
			var res = new GoStruct
			{
				Name = pc.Children[0].ToString().Replace("ID: ", "")
			};

			var root = pc.Children.First(x => x.ToString() == "anon_struct")
			    .Children.FirstOrDefault(x => x.ToString() == "struct_content");


			ParseStructHelp(root, res);

			//Console.WriteLine(JsonSerializer.Serialize(res));

			sw.WriteLine(JsonSerializer.Serialize(res));
		}

		static void ParseStructHelp(Node root, GoStruct res)
		{
			if (root == null) // empty struct
				return;

			bool onlyOneField = root.Children.Count <= 2;
			string lastType = "";
			string lastDelim = "";
			root.Children.Reverse();
			foreach (var item in root.Children)
			{
				if (item.ToString() == "struct_delim")
				{
					lastDelim = item.Children[0].ToString().Contains(",") ? "," : "\n";
					continue;
				}


				var goTypes = item.Children.Where(x => x.ToString() == "go_type");
				if (goTypes.Count() > 1 || lastDelim == "\n" || onlyOneField) // embeded structs
				{
					lastType = goTypes.Last().Children.First(x => x.ToString() != "arr_ptr").ToString().Replace("ID: ", "");
				}


				res.Types.Add(lastType);
			}
		}

		static void ParseMultilineType(Node pc, StreamWriter sw)
		{
			foreach (var pcc in pc.Children) // pcc -- line_type
			{
				var pccc = pcc.Children.FirstOrDefault(x => x.ToString() == "anon_struct");
				if (pccc == null)
					continue;

				var res = new GoStruct
				{
					Name = pcc.Children[0].ToString().Replace("ID: ", "")
				};

				ParseStructHelp(pccc.Children.FirstOrDefault(x => x.ToString() == "struct_content"), res);
				//Console.WriteLine(JsonSerializer.Serialize(res));
				sw.WriteLine(JsonSerializer.Serialize(res));
			}
		}


		static void ParseFunc(Node pc, StreamWriter sw)
		{
			var res = new GoFunc();

			foreach (var pcc in pc.Children)
			{
				var opt = pcc.ToString();

				switch (opt)
				{
					case "f_args":
						var args = pcc.Children.Where(x => x.ToString().StartsWith("f_arg", StringComparison.Ordinal));
						res.ArgsCnt = args.Count();
						if (res.ArgsCnt == 0)
							break;
						foreach (var arg in args)
						{
							res.Args.Add(arg.ToString().Replace("f_arg: ", ""));
						}
						break;
					case "f_returns":
						res.Return = pcc.Children.Count(x => x.ToString() == "f_return" || x.ToString().StartsWith("go_type", StringComparison.Ordinal));
						break;
					case "f_reciever":
						res.Receiver = pcc.Children.FirstOrDefault(x => x.ToString() == "f_type")?.Children[0]?.ToString()?.Replace("ID: ", "");
						if (res.Receiver == null) // highlevel grammar
						{
							var components = pcc.Children[0].ToString().Split(' ');
							var idx = components.Length - 1;
							for (var i = 0; i < components.Length; i++)
							{
								if (components[i] == "[")
								{
									idx = i - 1;
									break;
								}
							}
							res.Receiver = components[idx];
						}

						break;
					default:
						if (opt.StartsWith("f_name: ", StringComparison.Ordinal)) res.Name = opt.Replace("f_name: ", "");
						break;
				}
			}

			if (!res.Empty)
			{
				sw.WriteLine(JsonSerializer.Serialize(res));
			}
		}
		internal static void Serialize(string path, Node root)
		{
			FileInfo file = new FileInfo(path);
			file.Directory.Create();

			using (StreamWriter sw = File.CreateText(path))
			{
				foreach (var r in root.Children)
				{
					if (r.ToString() != "package_content")
						continue;

					foreach (var pc in r.Children)
					{
						if (pc.ToString() == "func")
							ParseFunc(pc, sw);

						if (pc.ToString() == "type_def")
						{
							var pcc = pc.Children[1];
							var c = pcc.ToString();

							switch (c)
							{
								case "struct_type":
									ParseStruct(pcc, sw);
									break;
								case "multiline_type":
									ParseMultilineType(pcc, sw);
									break;
							}
						}

					}
				}
			}

		}
	}
}
