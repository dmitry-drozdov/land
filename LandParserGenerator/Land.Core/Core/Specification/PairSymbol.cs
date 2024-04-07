using System;
using System.Collections.Generic;
using System.Linq;

namespace Land.Core.Specification
{
	[Serializable]
	public class PairSymbol: ISymbol
	{
		public string Name { get; set; }
		public string Left { get; set; }
		public string Right { get; set; }

		public override bool Equals(object obj)
		{
			return obj is PairSymbol symbol && symbol.Name == Name;
		}

		public override int GetHashCode()
		{
			return Name.GetHashCode();
		}

		public override string ToString()
		{
			return Name;
		}
	}
}
