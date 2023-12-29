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
                        if (pc.ToString() != "func")
                            continue;

                        var res = new GoParseResult();

                        foreach (var pcc in pc.Children)
                        {
                            var opt = pcc.ToString();

                            switch (opt)
                            {
                                case "f_name":
                                    res.Name = pcc.Children[0].ToString().Replace("ID: ", "");
                                    break;
                                case "f_args":
                                    var args = pcc.Children.Where(x => x.ToString().StartsWith("f_arg"));
                                    res.ArgsCnt = args.Count();
                                    if (res.ArgsCnt == 0)
                                        break;
                                    foreach (var arg in args)
                                    {
                                        res.Args.Add(arg.ToString().Replace("f_arg: ", ""));
                                    }
                                    break;
                                case "f_returns":
                                    res.Return = pcc.Children.Count(x => x.ToString() == "f_return" || x.ToString().StartsWith("go_type"));
                                    break;
                            }
                        }

                        if (!res.Empty)
                        {
                            sw.WriteLine(JsonSerializer.Serialize(res));
                        }
                    }
                }
            }

        }
    }
}
