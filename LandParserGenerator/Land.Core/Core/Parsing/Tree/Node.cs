using System;
using System.Collections.Generic;
using System.Linq;
using Land.Core.Specification;
using Land.Core.Lexing;

namespace Land.Core.Parsing.Tree
{
	[Serializable]
	public class Node
	{
		// Shared empty instances to avoid per-node allocations on read.
		// IMPORTANT: do not mutate these instances (use GetMutableOptions/GetMutableArguments if you need to write).
		private static readonly SymbolOptionsManager _emptyOptions = new SymbolOptionsManager();
		private static readonly SymbolArguments _emptyArguments = new SymbolArguments();

		// --------------------
		// Lazy / lightweight
		// --------------------

		// Lazy Guid: generated only if/when Id is accessed
		private Guid _id;
		public Guid Id
		{
			get
			{
				if (_id == Guid.Empty)
					_id = Guid.NewGuid();
				return _id;
			}
		}

		/// <summary>Родительский узел</summary>
		public Node Parent { get; set; }

		/// <summary>Символ грамматики, которому соответствует узел</summary>
		public string Symbol { get; set; }

		public string UserifiedSymbol { get; set; }

		/// <summary>
		/// Псевдоним символа, которому соответствует узел
		/// в соответствии с подструктурой
		/// </summary>
		public string Alias { get; set; }

		// --------------------
		// Lazy Any token range
		// --------------------
		private TokenStream _lazyTokenStream;
		private int _lazyStartIndex = -1;      // inclusive
		private int _lazyEndExclusive = -1;    // exclusive
		private bool _hasExplicitValue;

		public bool HasExplicitValue => _hasExplicitValue;

		public void SetLazyTokenRange(TokenStream stream, int startIndex, int endExclusive)
		{
			_lazyTokenStream = stream;
			_lazyStartIndex = startIndex;
			_lazyEndExclusive = endExclusive;
		}

		public bool TryGetLazyTokenRange(out TokenStream stream, out int startIndex, out int endExclusive)
		{
			if (_lazyTokenStream != null && _lazyStartIndex >= 0 && _lazyEndExclusive >= _lazyStartIndex)
			{
				stream = _lazyTokenStream;
				startIndex = _lazyStartIndex;
				endExclusive = _lazyEndExclusive;
				return true;
			}

			stream = null;
			startIndex = -1;
			endExclusive = -1;
			return false;
		}

		public void ExtendLazyTokenRangeEnd(int newEndExclusive)
		{
			if (_lazyTokenStream == null || _lazyStartIndex < 0 || _lazyEndExclusive < _lazyStartIndex)
				return;
			if (newEndExclusive <= _lazyEndExclusive)
				return;

			// If Value has already been materialized, keep it consistent by appending missing tokens.
			if (_value != null && _value.Count > 0)
			{
				for (int i = _lazyEndExclusive; i < newEndExclusive; i++)
					_value.Add(_lazyTokenStream.GetTokenAt(i).Text);
			}

			_lazyEndExclusive = newEndExclusive;
		}

		private void ClearLazyTokenRange()
		{
			_lazyTokenStream = null;
			_lazyStartIndex = -1;
			_lazyEndExclusive = -1;
		}


		/// <summary>
		/// Набор токенов, соответствующих листовому узлу
		/// (создаётся только если реально задавали значение)
		/// </summary>
		private List<string> _value;
		public List<string> Value
		{
			get
			{
				// If this node represents a lazy Any range, materialize on demand.
				if (_value == null && _lazyTokenStream != null && _lazyStartIndex >= 0 && _lazyEndExclusive >= _lazyStartIndex)
				{
					int cnt = _lazyEndExclusive - _lazyStartIndex;
					_value = new List<string>(cnt > 0 ? cnt : 1);
					for (int i = _lazyStartIndex; i < _lazyEndExclusive; i++)
						_value.Add(_lazyTokenStream.GetTokenAt(i).Text);
					_hasExplicitValue = true;
				}

				if (_value == null)
					_value = new List<string>(1);
				return _value;
			}
			set
			{
				_value = value;
				_hasExplicitValue = true;
				// If caller explicitly sets Value, lazy range is no longer authoritative.
				ClearLazyTokenRange();
			}
		}
		/// <summary>
		/// Потомки узла (создаётся только если реально добавляли детей)
		/// </summary>
		private List<Node> _children;
		public List<Node> Children
		{
			get
			{
				if (_children == null)
					_children = new List<Node>(2);
				return _children;
			}
			set => _children = value;
		}

		// ВАЖНО: Options/Arguments читаются очень часто в визиторах.
		// Чтобы чтение НЕ создавало новые объекты на каждом узле,
		// возвращаем shared empty по умолчанию и аллоцируем только при явном присваивании.
		private static readonly SymbolOptionsManager EmptyOptions = new SymbolOptionsManager(
			new Dictionary<string, Dictionary<string, List<dynamic>>>()
		);
		private static readonly SymbolArguments EmptyArguments = new SymbolArguments();

		/// <summary>
		/// Опции, связанные с конкретным вхождением символа,
		/// породившего данный узел
		/// </summary>
		private SymbolOptionsManager _options;
		public SymbolOptionsManager Options
		{
			get => _options ?? EmptyOptions;
			set => _options = value;
		}

		private SymbolArguments _arguments;
		public SymbolArguments Arguments
		{
			get => _arguments ?? EmptyArguments;
			set => _arguments = value;
		}

		public string Type => Alias ?? UserifiedSymbol ?? Symbol;

		protected SegmentLocation _location;
		private bool LocationReady { get; set; } = false;
		public SegmentLocation Location
		{
			get
			{
				if (!LocationReady)
					GetLocationFromChildren();
				return _location;
			}
		}

		public Node(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null)
		{
			Symbol = symbol;
			_options = opts;
			_arguments = args;
		}

		public Node(Node node)
		{
			_id = node._id;
			Symbol = node.Symbol;
			UserifiedSymbol = node.UserifiedSymbol;
			_options = node._options;
			_arguments = node._arguments;
			Parent = node.Parent;
			Alias = node.Alias;
			_children = node._children;
			_value = node._value;

			_lazyTokenStream = node._lazyTokenStream;
			_lazyStartIndex = node._lazyStartIndex;
			_lazyEndExclusive = node._lazyEndExclusive;
			_hasExplicitValue = node._hasExplicitValue;
			_location = node._location;
			LocationReady = node.LocationReady;
		}

		public void CopyFromNode(Node node)
		{
			_id = node._id;
			Symbol = node.Symbol;
			UserifiedSymbol = node.UserifiedSymbol;
			_options = node._options;
			_arguments = node._arguments;
			Parent = node.Parent;
			Alias = node.Alias;
			_children = node._children;
			_value = node._value;

			_lazyTokenStream = node._lazyTokenStream;
			_lazyStartIndex = node._lazyStartIndex;
			_lazyEndExclusive = node._lazyEndExclusive;
			_hasExplicitValue = node._hasExplicitValue;
			_location = node._location;
			LocationReady = node.LocationReady;
		}

		protected void GetLocationFromChildren()
		{
			if (Children.Count > 0)
			{
				_location = Children[0].Location;

				foreach (var child in Children)
				{
					if (child.Location == null)
						child.GetLocationFromChildren();

					if (_location == null)
						_location = child.Location;
					else
						_location = _location.SmartMerge(child.Location);
				}
			}

			LocationReady = true;
		}

		/// <summary>
		/// Возвращает текст токенов из области, соответствующей данному узлу
		/// </summary>
		public List<string> GetValue()
		{
			if (_value != null && _value.Count > 0)
				return new List<string>(_value);

			return Children.SelectMany(c => c.GetValue()).ToList();
		}

		public void AddLastChild(Node child, bool mergeLocation = false)
		{
			Children.Add(child);
			child.Parent = this;

			if (mergeLocation)
			{
				_location = Location != null
					? Location.SmartMerge(child.Location)
					: child.Location;
			}
			else
			{
				ResetLocation();
			}
		}

		public void ReplaceChild(Node child, int position, bool mergeLocation = false)
		{
			if (position <= Children.Count)
			{
				Children.RemoveAt(position);
				InsertChild(child, position, mergeLocation);
			}
		}

		public void InsertChild(Node child, int position, bool mergeLocation = false)
		{
			if (position <= Children.Count)
			{
				if (position == Children.Count)
					AddLastChild(child);
				else
				{
					Children.Insert(position, child);
					child.Parent = this;

					if (mergeLocation)
					{
						_location = Location != null
							? Location.SmartMerge(child.Location)
							: child.Location;
					}
					else
					{
						ResetLocation();
					}
				}
			}
		}

		public void AddFirstChild(Node child, bool mergeLocation = false)
		{
			Children.Insert(0, child);
			child.Parent = this;

			if (mergeLocation)
			{
				_location = Location != null
					? Location.SmartMerge(child.Location)
					: child.Location;
			}
			else
			{
				ResetLocation();
			}
		}

		public void ResetChildren()
		{
			if (_children != null)
				_children.Clear();
			ResetLocation();
		}

		public void ResetLocation()
		{
			_location = null;
			LocationReady = false;
		}

		public void Reset()
		{
			ResetChildren();
			if (_value != null)
				_value.Clear();
			_hasExplicitValue = false;
			ClearLazyTokenRange();
		}
		public void SetLocation(PointLocation start, PointLocation end)
		{
			_location = new SegmentLocation()
			{
				Start = start,
				End = end
			};

			LocationReady = true;
		}

		public void SetValue(params string[] vals)
		{
			if (_value == null)
				_value = new List<string>(vals.Length);
			else
				_value.Clear();

			_value.AddRange(vals);
			_hasExplicitValue = true;
			ClearLazyTokenRange();
		}
		public virtual void Accept(BaseTreeVisitor visitor)
		{
			visitor.Visit(this);
		}

		public override string ToString()
		{
			return (String.IsNullOrEmpty(Alias) ? UserifiedSymbol ?? Symbol : Alias)
				+ (_value != null && _value.Count > 0 ? ": " + String.Join(" ", _value.Select(v => v.Trim())) : "");
		}
	}
}
