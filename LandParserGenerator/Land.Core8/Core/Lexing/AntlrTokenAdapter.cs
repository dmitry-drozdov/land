using System;
using System.Collections.Generic;
using System.Linq;
using System.IO;
using System.Diagnostics;
using System.Text;
using Antlr4.Runtime;
using System.Xml.Linq;

namespace Land.Core.Lexing
{
	public class AntlrTokenAdapter : IToken
	{
		private Antlr4.Runtime.IToken Token { get; set; }

		public SegmentLocation Location { get; private set; }
		public string Text
		{
			get
			{
				if (text == null)
					text = Token.Text;
				return text;
			}
		}
		public string Name { get; private set; }
		public int Type { get; private set; }

		private string text = null;

		public AntlrTokenAdapter(Antlr4.Runtime.IToken token, Antlr4.Runtime.Lexer lexer)
		{
			Token = token;
			Type = Token.Type;
			Name = lexer.Vocabulary.GetSymbolicName(Token.Type);

			Location = new SegmentLocation()
			{
				Start = new PointLocation(Token.Line, Token.Column, Token.StartIndex),
				End = new PointLocation(Token.StopIndex)
			};
		}

	}
}
