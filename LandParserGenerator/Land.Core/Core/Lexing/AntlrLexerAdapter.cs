using System;
using System.Collections.Generic;
using System.Linq;
using System.IO;
using System.Text;
using Antlr4.Runtime;

namespace Land.Core.Lexing
{
	public class AntlrLexerAdapter : ILexer
	{
		private Lexer Lexer { get; set; }

		private Func<ICharStream, Lexer> LexerConstructor { get; set; }

		public AntlrLexerAdapter(Func<ICharStream, Lexer> constructor)
		{
			LexerConstructor = constructor;
		}

		public void SetSourceFile(string filename)
		{
			var stream = new UnbufferedCharStream(new StreamReader(filename, Encoding.Default, true));
			Lexer = LexerConstructor(stream);
		}

		public void SetSourceText(string text)
		{
			/*byte[] textBuffer = Encoding.UTF8.GetBytes(text);
			MemoryStream memStream = new MemoryStream(textBuffer);*/

			var stream = CharStreams.fromstring(text);

			Lexer = LexerConstructor(stream);
		}

		public IToken NextToken()
		{
			return new AntlrTokenAdapter(Lexer.NextToken(), Lexer);
		}

		public IToken CreateToken(string name, int type)
		{
			return new StubToken(name, type);
		}

		public IList<Antlr4.Runtime.IToken> GetAllTokens()
		{
			var res = new List<Antlr4.Runtime.IToken>();
			var t = Lexer.NextToken();
			while (t.Type != -1)
			{
				res.Add(t);
				t = Lexer.NextToken();
			}
			return res;
		}

		/*public IList<AntlrTokenAdapter> GetAllTokens()
	{
		var res = new List<AntlrTokenAdapter>();
		var t = new AntlrTokenAdapter(Lexer.NextToken(), Lexer);
		while (t.Type != -1)
		{
			res.Add(t);
			t = new AntlrTokenAdapter(Lexer.NextToken(), Lexer);
		}
		return res;
	}*/

	}
}
