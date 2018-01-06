%{
	public List<string> Log = new List<string>();
%}

%using System.Linq;
%using QUT.Gppg;

%namespace LandParserGenerator.Builder

%option stack

%x in_terminal_declaration
%x in_options
%x in_skip
%x in_regex

LETTER [_a-zA-Z]
DIGIT [0-9]
NUMBER {DIGIT}+
ID {LETTER}({LETTER}|{DIGIT})*

LINE_COMMENT "//".*    
MULTILINE_COMMENT "/*"([^*]|\*[^/])*"*/"
STRING \'([^'\\]*|(\\\\)+|\\[^\\])*\'

%%

%{
  yylloc = new LexLocation(tokLin, tokCol, tokELin, tokECol);
%}

{LINE_COMMENT} |
{MULTILINE_COMMENT} {}

// Группа и все возможные квантификаторы

"(" return (int)Tokens.LPAR;

")" return (int)Tokens.RPAR;

"+" {
	yylval.quantVal = Quantifier.ONE_OR_MORE; 
	return (int)Tokens.ONE_OR_MORE;
}

"*" {
	yylval.quantVal = Quantifier.ZERO_OR_MORE; 
	return (int)Tokens.ZERO_OR_MORE;
}

"?" {
	yylval.quantVal = Quantifier.ZERO_OR_ONE;
	return (int)Tokens.OPTIONAL;
}

// Начало правила

^{ID} {
	yylval.strVal = yytext;
	return (int)Tokens.ENTITY_NAME;
}

"=" return (int)Tokens.EQUALS;

// Символы, означающие нечто внутри правила

"|" return (int)Tokens.OR;

"~" return (int)Tokens.IS_LIST_NODE;

// Элементы правила

{ID} {
	yylval.strVal = yytext;
	return (int)Tokens.ID;
}

{STRING} {
	yylval.strVal = yytext;
	return (int)Tokens.STRING;
}

^{ID}[" "\t]*: {
	yylval.strVal = yytext.Split(' ')[0].TrimEnd(':').Trim();
	BEGIN(in_terminal_declaration);
	return (int)Tokens.ENTITY_NAME;
}

<in_terminal_declaration> {
	.+[\n\r] {
		BEGIN(0);
		
		yylval.strVal = yytext.Trim();
		return (int)Tokens.REGEX;
	}
}

^"%%" {
	BEGIN(in_options);
	return (int)Tokens.PROC;
}

<in_options> {
	^"%"{ID} {
		var optionName = yytext.ToLower().Trim('%');
		switch(optionName)
		{
			case "start":
				return (int)Tokens.OPTION_START;
			case "skip":
				return (int)Tokens.OPTION_SKIP;
			case "ghost":
				return (int)Tokens.OPTION_GHOST;
			default:
				return (int)Tokens.error;
		}
	}

	{ID} {
		yylval.strVal = yytext;
		return (int)Tokens.ID;
	}
	
	"." {
		return (int)Tokens.DOT;
	}
	
	{LINE_COMMENT} |
	{MULTILINE_COMMENT} {}
}

%%

public override void yyerror(string format, params object[] args)
{ 
	Log.Add(String.Format(format, args.Select(a=>a.ToString())));
}