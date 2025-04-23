using System;
using System.Collections.Generic;
using System.Linq;

namespace Land.Core.Specification
{
	public enum AnyArgument { Include, Except, Avoid, IgnorePairs, Error }

	[Serializable]
	public class SymbolArguments
	{
		#region Any

		public Dictionary<AnyArgument, HashSet<string>> AnyArguments { get; set; } =
			new Dictionary<AnyArgument, HashSet<string>>();

		public void Set(AnyArgument anyArg, IEnumerable<string> symbols) =>
			AnyArguments[anyArg] = new HashSet<string>(symbols);

		public bool Contains(AnyArgument anyArg) =>
			AnyArguments.ContainsKey(anyArg);

		public bool Contains(AnyArgument anyArg, string token) =>
			AnyArguments.ContainsKey(anyArg) && AnyArguments[anyArg].Contains(token);

		public Dictionary<AnyArgument, List<List<string>>> AnyArgumentsList { get; set; } =
			new Dictionary<AnyArgument, List<List<string>>>();

		public void SetList(AnyArgument anyArg, List<List<string>> symbols) =>
			AnyArgumentsList[anyArg] = new List<List<string>>(symbols);

		public bool ContainsList(AnyArgument anyArg) =>
			AnyArgumentsList.ContainsKey(anyArg);

		public bool ContainsList(AnyArgument anyArg, List<string> tokens) =>
			AnyArgumentsList.ContainsKey(anyArg) && AnyArgumentsList[anyArg].Contains(tokens);

		#endregion

		public SymbolArguments Clone() =>
			new SymbolArguments()
			{
				AnyArguments = AnyArguments.ToDictionary(e => e.Key, e => new HashSet<string>(e.Value))
			};

		public override bool Equals(object obj)
		{
			/// Если у обоих наборов аргументов ключи одинаковые
			if (obj is SymbolArguments args
				&& this.AnyArguments.Keys.SequenceEqual(args.AnyArguments.Keys))
			{
				/// Проверяем множества символов, связанные с каждым ключом
				foreach (var key in this.AnyArguments.Keys)
				{
					if (!args.AnyArguments[key].SequenceEqual(this.AnyArguments[key]))
					{
						return false;
					}
				}

				return true;
			}
			else
			{
				return false;
			}
		}
	}
}
