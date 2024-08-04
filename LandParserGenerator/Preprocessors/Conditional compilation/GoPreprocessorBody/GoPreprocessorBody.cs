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

namespace GoPreprocessingBody.ConditionalCompilation
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
			RemoveIncorrectCalls(root);
		}

		void RemoveIncorrectCalls(Node root)
		{
			if (root == null)
				return;

			for (var i = 0; i < root.Children.Count; i++)
			{
				var child = root.Children[i];
				RemoveIncorrectCalls(child);
				if (child.ToString() == "call" && child.Children[2].Children.Count <= 1)
				{
					var name = child.Children[0].ToString().Remove(0, 4);
					if (name == "int" || name == "int32" || name == "int64" ||
						name == "uint" || name == "uint32" || name == "uint64" ||
						name == "float32" || name == "float64" || name == "string")
					{
						root.Children.RemoveAt(i);
					}
				}
			}
		}
	}
}
