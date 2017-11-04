﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.IO;
using System.Text;
using System.Threading.Tasks;

namespace LandParserGenerator.Parsing.LR
{
	/// <summary>
	/// Таблица LR(1) парсинга
	/// </summary>
	public class TableLR1
	{
		private HashSet<Action>[,] Table { get; set; }
		private Dictionary<string, int> Lookaheads { get; set; }
		private Dictionary<string, int> NonterminalSymbols { get; set; }

		/// <summary>
		/// Множества состояний (множество множеств пунктов)
		/// </summary>
		private List<HashSet<Marker>> Items { get; set; }
		/// <summary>
		/// Переходы между состояниями
		/// </summary>
		private List<Dictionary<string, int>> Transitions { get; set; }

		public TableLR1(Grammar g)
		{
			Lookaheads = g.Tokens.Keys
				.Zip(Enumerable.Range(0, g.Tokens.Count), (a, b) => new { smb = a, idx = b })
				.ToDictionary(e => e.smb, e => e.idx);

			/// Строим набор множеств пунктов
			BuildItems(g);

			Table = new HashSet<Action>[Items.Count, Lookaheads.Count];

			for(var i=0; i<Items.Count;++i)
			{
				foreach (var lookahead in Lookaheads)
					this[i, lookahead.Key] = new HashSet<Action>();

				foreach(var marker in Items[i])
				{
					/// A => alpha * a beta
					if(g[marker.Next] is TerminalSymbol)
					{
						this[i, marker.Next].Add(new ShiftAction()
						{
							TargetItemIndex = Transitions[i][marker.Next]
						});
					}

					/// A => alpha *
					if (String.IsNullOrEmpty(marker.Next) 
						&& marker.Alternative.NonterminalSymbolName != g.StartSymbol)
					{
						this[i, marker.Lookahead].Add(new ReduceAction()
						{
							ReductionAlternative = marker.Alternative
						});
					}
				}

				/// S => ...*, $
				if (Items[i].Any(item=>item.Alternative.NonterminalSymbolName == g.StartSymbol 
					&& String.IsNullOrEmpty(item.Next)
					&& item.Lookahead == Grammar.EOF_TOKEN_NAME))
				{
					this[i, Grammar.EOF_TOKEN_NAME].Add(new AcceptAction());
				}
			}
		}

		private void BuildItems(Grammar g)
		{
			Items = new List<HashSet<Marker>>()
			{
				g.BuildClosure(new HashSet<Marker>(
					(g[g.StartSymbol] as NonterminalSymbol).Alternatives.Select(a=>new Marker(a, 0, Grammar.EOF_TOKEN_NAME))
				))
			};

			Transitions = new List<Dictionary<string, int>>();

			for (var i = 0; i < Items.Count; ++i)
			{
				Transitions.Add(new Dictionary<string, int>());

				foreach (var smb in g.Tokens.Keys.Union(g.Rules.Keys))
				{
					var gotoSet = g.Goto(Items[i], smb);

					if (gotoSet.Count > 0)
					{
						/// Проверяем, не совпадает ли полученное множество 
						/// с каким-либо из имеющихся
						var j = 0;
						for (; j < Items.Count; ++j)
							if (EqualMarkerSets(Items[j], gotoSet))
							{
								break;
							}

						/// Если не нашли совпадение
						if (j == Items.Count)
						{
							Items.Add(gotoSet);
						}

						Transitions[i][smb] = j;
					}
				}
			}
		}

		private bool EqualMarkerSets(HashSet<Marker> a, HashSet<Marker> b)
		{
			if (a.Count != b.Count)
				return false;

			foreach(var elem in a)
			{
				if (!b.Contains(elem))
					return false;
			}

			return true;
		}

		public HashSet<Action> this[int i, string lookahead]
		{
			get { return Table[i, Lookaheads[lookahead]]; }

			private set { Table[i, Lookaheads[lookahead]] = value; }
		}

		public void ExportToCsv(string filename)
		{
			//var output = new StreamWriter(filename);

			//var orderedLookaheads = Lookaheads.OrderBy(l => l.Value);
			//output.WriteLine("," + String.Join(",", orderedLookaheads.Select(l => l.Key)));

			//foreach (var nt in NonterminalSymbols.Keys)
			//{
			//	output.Write($"{nt},");

			//	output.Write(String.Join(",",
			//		orderedLookaheads.Select(l=>this[nt, l.Key])
			//		.Select(alts => alts.Count == 0 ? "" : alts.Count == 1 ? alts.Single().ToString() : String.Join("/", alts))));

			//	output.WriteLine();
			//}

			//output.Close();
		}
	}
}