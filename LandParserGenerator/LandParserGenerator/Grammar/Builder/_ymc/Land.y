%{
    public Parser(AbstractScanner<LandParserGenerator.Builder.ValueType, LexLocation> scanner) : base(scanner) { }
    
    public Grammar ConstructedGrammar;
%}

%using System.Linq;
%using LandParserGenerator;

%output = LandParser.cs

%namespace LandParserGenerator.Builder

%union { 
	public int intVal; 
	public bool boolVal;
	public string strVal;
	public List<string> strList;
	
	public List<Alternative> altList;
	// Информация о количестве повторений
	public Nullable<Quantifier> quantVal;
}

%start lp_description

%left OR
%token COLON LPAR RPAR COMMA PROC EQUALS MINUS PLUS EXCLAMATION ADD_CHILD DOT
%token <strVal> REGEX NAMED STRING ID ENTITY_NAME
%token <intVal> POSITION
%token <quantVal> OPTIONAL ZERO_OR_MORE ONE_OR_MORE
%token OPTION_SKIP OPTION_GHOST OPTION_START IS_LIST_NODE

%type <quantVal> quantifier
%type <strVal> body_element_core body_element_atom group body_element
%type <strList> identifiers
%type <altList> body
%type <boolVal> is_list_node

%%

lp_description 
	: structure PROC options
	;

/***************************** STRUCTURE ******************************/
	
structure 
	: structure element
	| element
	;

element
	: terminal
	| nonterminal
	;
	
terminal
	: ENTITY_NAME REGEX { ConstructedGrammar.DeclareTerminal($1, $2); }
	;

/******* ID = ID 'string' (group)[*|+|?]  ********/
nonterminal
	: ENTITY_NAME EQUALS body 
		{ ConstructedGrammar.DeclareNonterminal($1, $3); }
	;
	
body
	: body body_element 
		{ 
			$$ = $1; 
			$$[$$.Count-1].Add($2); 	
		}
	| body OR 
		{ 
			$$ = $1;
			$$.Add(new Alternative());		
		}
	|  
		{ 
			$$ = new List<Alternative>(); 
			$$.Add(new Alternative()); 
		}
	;
	
body_element
	: is_list_node body_element_core quantifier 
		{ 
			if($3.HasValue)
			{
				var generated = ConstructedGrammar.GenerateNonterminal($2, $3.Value);
				$$ = new Entry(generated);
				
				if($1) { ConstructedGrammar.SetListSymbol($$); }
			}
			else
			{
				$$ = new Entry($2);
			}
		}
	;
	
is_list_node
	: IS_LIST_NODE { $$ = true; }
	| { $$ = false; }
	;
	
quantifier
	: OPTIONAL { $$ = $1; }
	| ZERO_OR_MORE { $$ = $1; }
	| ONE_OR_MORE { $$ = $1; }
	| { $$ = null; }
	;
	
body_element_core
	: body_element_atom
		{ $$ = $1; }
	| group 
		{ $$ = $1; }
	;
	
body_element_atom
	: STRING
		{ 
			$$ = ConstructedGrammar.GenerateTerminal($1);
		}
	| ID 
		{ $$ = $1; }
	;
	
group
	: LPAR body RPAR { $$ = ConstructedGrammar.GenerateNonterminal($2); }
	;

/***************************** OPTIONS ******************************/

options
	:
	| options option
	;
	
option
	: ghost_option
	| skip_option
	| start_option
	;
	
skip_option
	: OPTION_SKIP identifiers
		{ ConstructedGrammar.SetSkipTokens($2.ToArray()); }
	;	

ghost_option
	: OPTION_GHOST identifiers
		{ ConstructedGrammar.SetGhostSymbols($2.ToArray()); }
	;
	
start_option
	: OPTION_START ID
		{ ConstructedGrammar.SetStartSymbol($2); }
	;
	
identifiers
	: identifiers ID 
		{ $$ = $1; $$.Add($2); }
	| ID 
		{ $$ = new List<string>(); $$.Add($1); }
	;
