using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Land.Core.Lexing
{
	public interface ILexer
	{
		AntlrTokenAdapter NextToken();

		AntlrTokenAdapter CreateToken(string name, int type);

		void SetSourceFile(string filename);

		void SetSourceText(string text);
	}

}
