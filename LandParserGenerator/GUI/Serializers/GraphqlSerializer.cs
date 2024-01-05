using Land.Core.Parsing.Tree;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Security.Policy;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows.Shapes;

namespace Land.GUI.Serializers
{
    internal class GraphqlSerializer
    {
        GraphqlParseResult res = new GraphqlParseResult();

        //Parse this -> [[[[AgreementInput]!]!]]
        string ParseType(Node root)
        {
            var children = root.Children;
            if (children.Count == 0) return "";
            var result = "";
            foreach (var child in children)
            {
                var c = child.ToString();
                if (c == "type")
                {
                    result += ParseType(child);
                    continue;
                }
                if (c == "id")
                {
                    result += DecodeID(child.Children[0]); ;
                    continue;
                }
                var strs = c.Split(new string[] { ": " }, StringSplitOptions.None);
                if (strs[0] == "LSB" || strs[0] == "RSB" || strs[0] == "EXM") result += strs[1];
            }
            return result;
        }
        // Parse this -> agreements: [[[[AgreementInput]!]!]]
        GraphqlDef ParseTypeLine(Node root)
        {
            var result = new GraphqlDef();
            foreach (var pcc in root.Children)
            {
                var opt = pcc.ToString();

                switch (opt)
                {
                    case "id":
                        result.Name = DecodeID(pcc.Children[0]);
                        break;
                    case "type":
                        result.Type = ParseType(pcc);
                        break;
                }
            }
            return result;
        }
        string DecodeID(Node node)
        {
            return node.ToString().Split(new string[] { ": " }, StringSplitOptions.None)[1];
        }
        internal void Serialize(Node root)
        {
            foreach (var r in root.Children)
            {
                if (r.ToString() != "type_def")
                    continue;

                var typeDef = new GraphqlType();

                foreach (var pc in r.Children)
                {
                    if (pc.ToString() == "id")
                    {
                        typeDef.Name = DecodeID(pc.Children[0]);
                    }

                    if (pc.ToString() == "func_line")
                    {
                        var func = new GraphqlFunc
                        {
                            Parent = typeDef.Name // must be already filled 
                        };
                        foreach (var pcc in pc.Children)
                        {
                            var opt = pcc.ToString();

                            switch (opt)
                            {
                                case "id":
                                    func.Name = DecodeID(pcc.Children[0]);
                                    break;
                                case "func_arg":
                                    var arg = pcc.Children.First(x => x.ToString() == "type_line");
                                    func.Args.Add(ParseTypeLine(arg));
                                    break;
                                case "type": // return value
                                    func.Return = ParseType(pcc);
                                    break;
                            }
                        }
                        res.AddFunc(func);
                    }

                    if (pc.ToString() == "type_line")
                    {
                        typeDef.Defs.Add(ParseTypeLine(pc));
                    }
                }

                if (typeDef.Defs.Count == 0) continue;

                if (r.Children[0].ToString().Contains("input"))
                {
                    GraphqlInput inputDef = new GraphqlInput(typeDef);
                    res.AddInput(inputDef);
                }
                else
                {
                    res.AddType(typeDef);
                }
            }
        }

        internal void Dump(string path)
        {
            FileInfo file = new FileInfo(path);
            file.Directory.Create();

            using (StreamWriter sw = File.CreateText(path))
            {
                sw.WriteLine(JsonSerializer.Serialize(res));
            }
        }
    }
}
