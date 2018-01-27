// This code was generated by the Gardens Point Parser Generator
// Copyright (c) Wayne Kelly, John Gough, QUT 2005-2014
// (see accompanying GPPGcopyright.rtf)

// GPPG version 1.5.2
// Machine:  DESKTOP-QMIGNCH
// DateTime: 23.01.2018 18:45:42
// UserName: Алексей
// Input file <./Land.y - 23.01.2018 18:45:42>

// options: no-lines gplex

using System;
using System.Collections.Generic;
using System.CodeDom.Compiler;
using System.Globalization;
using System.Text;
using QUT.Gppg;
using System.Linq;
using LandParserGenerator;

namespace LandParserGenerator.Builder
{
public enum Tokens {error=2,EOF=3,OR=4,COLON=5,LPAR=6,
    RPAR=7,COMMA=8,PROC=9,EQUALS=10,MINUS=11,PLUS=12,
    EXCLAMATION=13,ADD_CHILD=14,DOT=15,REGEX=16,NAMED=17,STRING=18,
    ID=19,ENTITY_NAME=20,OPTION_NAME=21,POSITION=22,OPTIONAL=23,ZERO_OR_MORE=24,
    ONE_OR_MORE=25,IS_LIST_NODE=26,PREC_NONEMPTY=27};

public struct ValueType
{ 
	public int intVal; 
	public bool boolVal;
	public string strVal;
	public List<string> strList;
	
	public List<Alternative> altList;
	// ÐÐ½ÑÐ¾ÑÐ¼Ð°ÑÐ¸Ñ Ð¾ ÐºÐ¾Ð»Ð¸ÑÐµÑÑÐ²Ðµ Ð¿Ð¾Ð²ÑÐ¾ÑÐµÐ½Ð¸Ð¹
	public Nullable<Quantifier> quantVal;
}
// Abstract base class for GPLEX scanners
[GeneratedCodeAttribute( "Gardens Point Parser Generator", "1.5.2")]
public abstract class ScanBase : AbstractScanner<ValueType,LexLocation> {
  private LexLocation __yylloc = new LexLocation();
  public override LexLocation yylloc { get { return __yylloc; } set { __yylloc = value; } }
  protected virtual bool yywrap() { return true; }
}

// Utility class for encapsulating token information
[GeneratedCodeAttribute( "Gardens Point Parser Generator", "1.5.2")]
public class ScanObj {
  public int token;
  public ValueType yylval;
  public LexLocation yylloc;
  public ScanObj( int t, ValueType val, LexLocation loc ) {
    this.token = t; this.yylval = val; this.yylloc = loc;
  }
}

[GeneratedCodeAttribute( "Gardens Point Parser Generator", "1.5.2")]
public class Parser: ShiftReduceParser<ValueType, LexLocation>
{
  // Verbatim content from ./Land.y - 23.01.2018 18:45:42
    public Parser(AbstractScanner<LandParserGenerator.Builder.ValueType, LexLocation> scanner) : base(scanner) { }
    
    public Grammar ConstructedGrammar;
    public List<Message> Errors = new List<Message>();
  // End verbatim content from ./Land.y - 23.01.2018 18:45:42

#pragma warning disable 649
  private static Dictionary<int, string> aliases;
#pragma warning restore 649
  private static Rule[] rules = new Rule[31];
  private static State[] states = new State[39];
  private static string[] nonTerms = new string[] {
      "lp_description", "quantifier", "body_element_core", "body_element_atom", 
      "group", "body_element", "context_option", "identifiers", "body", "prec_nonempty", 
      "$accept", "structure", "options", "element", "terminal", "nonterminal", 
      "option", };

  static Parser() {
    states[0] = new State(new int[]{20,14},new int[]{-1,1,-12,3,-14,38,-15,13,-16,37});
    states[1] = new State(new int[]{3,2});
    states[2] = new State(-1);
    states[3] = new State(new int[]{9,4,20,14},new int[]{-14,12,-15,13,-16,37});
    states[4] = new State(-26,new int[]{-13,5});
    states[5] = new State(new int[]{21,7,3,-2},new int[]{-17,6});
    states[6] = new State(-27);
    states[7] = new State(new int[]{19,8});
    states[8] = new State(new int[]{19,11},new int[]{-8,9});
    states[9] = new State(new int[]{19,10,21,-28,3,-28});
    states[10] = new State(-29);
    states[11] = new State(-30);
    states[12] = new State(-3);
    states[13] = new State(-5);
    states[14] = new State(new int[]{5,15,10,17});
    states[15] = new State(new int[]{16,16});
    states[16] = new State(-7);
    states[17] = new State(-11,new int[]{-9,18});
    states[18] = new State(new int[]{4,20,21,36,9,-8,20,-8,18,-14,19,-14,6,-14},new int[]{-6,19,-7,21});
    states[19] = new State(-9);
    states[20] = new State(-10);
    states[21] = new State(new int[]{18,30,19,31,6,33},new int[]{-3,22,-4,29,-5,32});
    states[22] = new State(new int[]{23,26,24,27,25,28,27,-20,4,-20,21,-20,18,-20,19,-20,6,-20,9,-20,20,-20,7,-20},new int[]{-2,23});
    states[23] = new State(new int[]{27,25,4,-16,21,-16,18,-16,19,-16,6,-16,9,-16,20,-16,7,-16},new int[]{-10,24});
    states[24] = new State(-12);
    states[25] = new State(-15);
    states[26] = new State(-17);
    states[27] = new State(-18);
    states[28] = new State(-19);
    states[29] = new State(-21);
    states[30] = new State(-23);
    states[31] = new State(-24);
    states[32] = new State(-22);
    states[33] = new State(-11,new int[]{-9,34});
    states[34] = new State(new int[]{7,35,4,20,21,36,18,-14,19,-14,6,-14},new int[]{-6,19,-7,21});
    states[35] = new State(-25);
    states[36] = new State(-13);
    states[37] = new State(-6);
    states[38] = new State(-4);

    for (int sNo = 0; sNo < states.Length; sNo++) states[sNo].number = sNo;

    rules[1] = new Rule(-11, new int[]{-1,3});
    rules[2] = new Rule(-1, new int[]{-12,9,-13});
    rules[3] = new Rule(-12, new int[]{-12,-14});
    rules[4] = new Rule(-12, new int[]{-14});
    rules[5] = new Rule(-14, new int[]{-15});
    rules[6] = new Rule(-14, new int[]{-16});
    rules[7] = new Rule(-15, new int[]{20,5,16});
    rules[8] = new Rule(-16, new int[]{20,10,-9});
    rules[9] = new Rule(-9, new int[]{-9,-6});
    rules[10] = new Rule(-9, new int[]{-9,4});
    rules[11] = new Rule(-9, new int[]{});
    rules[12] = new Rule(-6, new int[]{-7,-3,-2,-10});
    rules[13] = new Rule(-7, new int[]{21});
    rules[14] = new Rule(-7, new int[]{});
    rules[15] = new Rule(-10, new int[]{27});
    rules[16] = new Rule(-10, new int[]{});
    rules[17] = new Rule(-2, new int[]{23});
    rules[18] = new Rule(-2, new int[]{24});
    rules[19] = new Rule(-2, new int[]{25});
    rules[20] = new Rule(-2, new int[]{});
    rules[21] = new Rule(-3, new int[]{-4});
    rules[22] = new Rule(-3, new int[]{-5});
    rules[23] = new Rule(-4, new int[]{18});
    rules[24] = new Rule(-4, new int[]{19});
    rules[25] = new Rule(-5, new int[]{6,-9,7});
    rules[26] = new Rule(-13, new int[]{});
    rules[27] = new Rule(-13, new int[]{-13,-17});
    rules[28] = new Rule(-17, new int[]{21,19,-8});
    rules[29] = new Rule(-8, new int[]{-8,19});
    rules[30] = new Rule(-8, new int[]{19});
  }

  protected override void Initialize() {
    this.InitSpecialTokens((int)Tokens.error, (int)Tokens.EOF);
    this.InitStates(states);
    this.InitRules(rules);
    this.InitNonTerminals(nonTerms);
  }

  protected override void DoAction(int action)
  {
#pragma warning disable 162, 1522
    switch (action)
    {
      case 2: // lp_description -> structure, PROC, options
{ Errors.AddRange(ConstructedGrammar.CheckValidity()); }
        break;
      case 7: // terminal -> ENTITY_NAME, COLON, REGEX
{ 
			SafeGrammarAction(() => { 
				ConstructedGrammar.DeclareTerminal(ValueStack[ValueStack.Depth-3].strVal, ValueStack[ValueStack.Depth-1].strVal);
				ConstructedGrammar.AddAnchor(ValueStack[ValueStack.Depth-3].strVal, LocationStack[LocationStack.Depth-3]);
			}, LocationStack[LocationStack.Depth-3]);
		}
        break;
      case 8: // nonterminal -> ENTITY_NAME, EQUALS, body
{ 
			SafeGrammarAction(() => { 
				ConstructedGrammar.DeclareNonterminal(ValueStack[ValueStack.Depth-3].strVal, ValueStack[ValueStack.Depth-1].altList);
				ConstructedGrammar.AddAnchor(ValueStack[ValueStack.Depth-3].strVal, LocationStack[LocationStack.Depth-3]);
			}, LocationStack[LocationStack.Depth-3]);
		}
        break;
      case 9: // body -> body, body_element
{ 
			CurrentSemanticValue.altList = ValueStack[ValueStack.Depth-2].altList; 
			CurrentSemanticValue.altList[CurrentSemanticValue.altList.Count-1].Add(ValueStack[ValueStack.Depth-1].strVal); 	
		}
        break;
      case 10: // body -> body, OR
{ 
			CurrentSemanticValue.altList = ValueStack[ValueStack.Depth-2].altList;
			CurrentSemanticValue.altList.Add(new Alternative());		
		}
        break;
      case 11: // body -> /* empty */
{ 
			CurrentSemanticValue.altList = new List<Alternative>(); 
			CurrentSemanticValue.altList.Add(new Alternative()); 
		}
        break;
      case 12: // body_element -> context_option, body_element_core, quantifier, prec_nonempty
{ 		
			NodeOption option;
			
			if(!Enum.TryParse<NodeOption>(ValueStack[ValueStack.Depth-4].strVal.ToUpper(), out option))
			{
					Errors.Add(Message.Error(
						"ÐÐµÐ¸Ð·Ð²ÐµÑÑÐ½Ð°Ñ Ð¾Ð¿ÑÐ¸Ñ '" + ValueStack[ValueStack.Depth-4].strVal + "'",
						LocationStack[LocationStack.Depth-4].StartLine, LocationStack[LocationStack.Depth-4].StartColumn,
						"LanD"
					));
					option = NodeOption.NONE;
			}
			
			if(ValueStack[ValueStack.Depth-2].quantVal.HasValue)
			{
				var generated = ConstructedGrammar.GenerateNonterminal(ValueStack[ValueStack.Depth-3].strVal, ValueStack[ValueStack.Depth-2].quantVal.Value, ValueStack[ValueStack.Depth-1].boolVal);
				ConstructedGrammar.AddAnchor(generated, CurrentLocationSpan);
				
				CurrentSemanticValue.strVal = new Entry(generated, option);
			}
			else
			{
				CurrentSemanticValue.strVal = new Entry(ValueStack[ValueStack.Depth-3].strVal, option);
			}
		}
        break;
      case 13: // context_option -> OPTION_NAME
{ CurrentSemanticValue.strVal = ValueStack[ValueStack.Depth-1].strVal; }
        break;
      case 14: // context_option -> /* empty */
{ CurrentSemanticValue.strVal = NodeOption.NONE.ToString(); }
        break;
      case 15: // prec_nonempty -> PREC_NONEMPTY
{ CurrentSemanticValue.boolVal = true; }
        break;
      case 16: // prec_nonempty -> /* empty */
{ CurrentSemanticValue.boolVal = false; }
        break;
      case 17: // quantifier -> OPTIONAL
{ CurrentSemanticValue.quantVal = ValueStack[ValueStack.Depth-1].quantVal; }
        break;
      case 18: // quantifier -> ZERO_OR_MORE
{ CurrentSemanticValue.quantVal = ValueStack[ValueStack.Depth-1].quantVal; }
        break;
      case 19: // quantifier -> ONE_OR_MORE
{ CurrentSemanticValue.quantVal = ValueStack[ValueStack.Depth-1].quantVal; }
        break;
      case 20: // quantifier -> /* empty */
{ CurrentSemanticValue.quantVal = null; }
        break;
      case 21: // body_element_core -> body_element_atom
{ CurrentSemanticValue.strVal = ValueStack[ValueStack.Depth-1].strVal; }
        break;
      case 22: // body_element_core -> group
{ CurrentSemanticValue.strVal = ValueStack[ValueStack.Depth-1].strVal; }
        break;
      case 23: // body_element_atom -> STRING
{ 
			CurrentSemanticValue.strVal = ConstructedGrammar.GenerateTerminal(ValueStack[ValueStack.Depth-1].strVal);
			ConstructedGrammar.AddAnchor(CurrentSemanticValue.strVal, CurrentLocationSpan);
		}
        break;
      case 24: // body_element_atom -> ID
{ CurrentSemanticValue.strVal = ValueStack[ValueStack.Depth-1].strVal; }
        break;
      case 25: // group -> LPAR, body, RPAR
{ 
			CurrentSemanticValue.strVal = ConstructedGrammar.GenerateNonterminal(ValueStack[ValueStack.Depth-2].altList);
			ConstructedGrammar.AddAnchor(CurrentSemanticValue.strVal, CurrentLocationSpan);
		}
        break;
      case 28: // option -> OPTION_NAME, ID, identifiers
{
			switch(ValueStack[ValueStack.Depth-2].strVal)
			{
				case "skip":
					SafeGrammarAction(() => { ConstructedGrammar.SetSkipTokens(ValueStack[ValueStack.Depth-1].strList.ToArray()); }, LocationStack[LocationStack.Depth-3]);
					break;
				case "start":
					SafeGrammarAction(() => { ConstructedGrammar.SetStartSymbol(ValueStack[ValueStack.Depth-1].strList[0]); }, LocationStack[LocationStack.Depth-3]);
					break;
				case "ignorecase":
					ConstructedGrammar.IsCaseSensitive = false;
					break;
				default:
					NodeOption option;	
					if(!Enum.TryParse<NodeOption>(ValueStack[ValueStack.Depth-2].strVal.ToUpper(), out option))
					{
						Errors.Add(Message.Error(
							"ÐÐµÐ¸Ð·Ð²ÐµÑÑÐ½Ð°Ñ Ð¾Ð¿ÑÐ¸Ñ '" + ValueStack[ValueStack.Depth-2].strVal + "'",
							LocationStack[LocationStack.Depth-3].StartLine, LocationStack[LocationStack.Depth-3].StartColumn,
							"LanD"
						));
					}
					else
					{
						SafeGrammarAction(() => { ConstructedGrammar.SetSymbolsForOption(option, ValueStack[ValueStack.Depth-1].strList.ToArray()); }, LocationStack[LocationStack.Depth-3]);
					}				
					break;
			}
		}
        break;
      case 29: // identifiers -> identifiers, ID
{ CurrentSemanticValue.strList = ValueStack[ValueStack.Depth-2].strList; CurrentSemanticValue.strList.Add(ValueStack[ValueStack.Depth-1].strVal); }
        break;
      case 30: // identifiers -> ID
{ CurrentSemanticValue.strList = new List<string>(); CurrentSemanticValue.strList.Add(ValueStack[ValueStack.Depth-1].strVal); }
        break;
    }
#pragma warning restore 162, 1522
  }

  protected override string TerminalToString(int terminal)
  {
    if (aliases != null && aliases.ContainsKey(terminal))
        return aliases[terminal];
    else if (((Tokens)terminal).ToString() != terminal.ToString(CultureInfo.InvariantCulture))
        return ((Tokens)terminal).ToString();
    else
        return CharToString((char)terminal);
  }


private void SafeGrammarAction(Action action, LexLocation loc)
{
	try
	{
		action();
	}
	catch(IncorrectGrammarException ex)
	{
		Errors.Add(Message.Error(
			ex.Message,
			loc.StartLine, loc.StartColumn,
			"LanD"
		));
	}
}

}
}
