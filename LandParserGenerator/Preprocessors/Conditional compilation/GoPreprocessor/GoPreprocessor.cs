using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.IO;

using Land.Core;
using Land.Core.Parsing;
using Land.Core.Parsing.Tree;
using Land.Core.Parsing.Preprocessing;

using sharp_preprocessor;

namespace GoPreprocessing.ConditionalCompilation
{
    public class SharpPreprocessor : BasePreprocessor
    {
        public SharpPreprocessor()
        {

        }

        public override string Preprocess(string text, out bool success)
        {
            success = true;
            return text;
        }

        public override void Postprocess(Node root, List<Message> log)
        {
            if (root == null)
                return;

            foreach (var r in root.Children)
            {
                if (r.ToString() != "package_content")
                    continue;

                foreach (var pc in r.Children)
                {
                    if (pc.ToString() != "func")
                        continue;

                    foreach (var pcc in pc.Children)
                    {
                        var opt = pcc.ToString();

                        if (opt != "f_args")
                            continue;

                        var args = pcc.Children.Where(x => x.ToString() == "f_arg");
                        if (args.Count() == 0)
                            break;

                        var onlyTypes = args.All(a => a.Children.Count(x => x.ToString().StartsWith("ID: ")
                                                                        || x.ToString().StartsWith("go_type")) == 1);

                        args = args.Reverse();
                        string lastType = null;
                        foreach (var arg in args)
                        {
                            var types = arg.Children.LastOrDefault(x => x.ToString().StartsWith("go_type"));
                            Node type;
                            if (types != null)
                            {
                                type = types.Children.First(x => x.ToString() != "arr_ptr");
                            }
                            else
                            {
                                type = arg.Children.FirstOrDefault(x => x.ToString().StartsWith("ID: ") && onlyTypes);
                            }
                            if (arg.Children.Count(x => x.ToString().StartsWith("go_type")) == 1 && !onlyTypes)
                            {
                                type = null; // nullify type because it is ID 
                            }
                            if (type != null)
                            {
                                lastType = type.ToString().Replace("ID: ", "");
                            }
                            if (lastType != null)
                            {
                                arg.SetValue(lastType);
                            }
                        }
                        args = args.Reverse();
                    }
                }
            }
        }
    }
}
