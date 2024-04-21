using Antlr4.Runtime;

namespace Land.Core.Lexing
{
	public class AntlrTokenAdapter
	{
		//private IToken Token { get; set; }

		public SegmentLocation Location { get; private set; }
		public string Text { get; private set; }
		public string Name { get; set; }
		public int Type { get; private set; }


		public AntlrTokenAdapter(IToken Token, Lexer lexer)
		{
			//Token = token;
			Type = Token.Type;
			Name = lexer.Vocabulary.GetSymbolicName(Token.Type);
			Text=Token.Text;

			Location = new SegmentLocation()
			{
				Start = new PointLocation(Token.Line, Token.Column, Token.StartIndex),
				End = new PointLocation(Token.StopIndex)
			};
		}

		public AntlrTokenAdapter(string name, int type, Lexer lexer)
		{
			//Token = lexer.TokenFactory.Create(type, "");
			Type = type;
			Name = name;

			Location = new SegmentLocation()
			{
				Start = new PointLocation(0, 0, 0),
				End = new PointLocation(0, 0, 0)
			};
		}

	}
}
