using System;
using System.Collections.Generic;
using System.Linq;
using Land.Core.Specification;

namespace Land.Core.Parsing.Tree
{
	[Serializable]
	public class Node
	{
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

		/// <summary>
		/// Родительский узел
		/// </summary
		public Node Parent { get; set; }

		/// <summary>
		/// Символ грамматики, которому соответствует узел
		/// </summary>
		public string Symbol { get; set; }

		public string UserifiedSymbol { get; set; }

		/// <summary>
		/// Псевдоним символа, которому соответствует узел
		/// в соответствии с подструктурой
		/// </summary>
		public string Alias { get; set; }

		/// <summary>
		/// Набор токенов, соответствующих листовому узлу
		/// </summary>
		private List<string> _value;
		public List<string> Value
		{
			get
			{
				if (_value == null)
					_value = new List<string>(1);
				return _value;
			}
			set => _value = value;
		}

		/// <summary>
		/// Потомки узла
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

		/// <summary>
		/// Опции, связанные с конкретным вхождением в грамматику символа,
		/// породившего данный узел
		/// </summary>
		private SymbolOptionsManager _options;
		public SymbolOptionsManager Options
		{
			get
			{
				if (_options == null)
					_options = new SymbolOptionsManager();
				return _options;
			}
			set => _options = value;
		}

		private SymbolArguments _arguments;
		public SymbolArguments Arguments
		{
			get
			{
				if (_arguments == null)
					_arguments = new SymbolArguments();
				return _arguments;
			}
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

			_location = node._location;
			LocationReady = node.LocationReady;
		}

		public void CopyFromNode(Node node)
		{
			_id = node._id;

			this.Symbol = node.Symbol;
			this.UserifiedSymbol = node.UserifiedSymbol;
			this._options = node._options;
			this._arguments = node._arguments;
			this.Parent = node.Parent;
			this.Alias = node.Alias;
			this._children = node._children;
			this._value = node._value;

			this._location = node._location;
			this.LocationReady = node.LocationReady;
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
		/// Возвращает текст токенов из области, 
		/// соответствующей данному узлу
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
