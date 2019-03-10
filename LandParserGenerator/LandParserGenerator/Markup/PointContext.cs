﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Runtime.Serialization;

using Land.Core.Parsing.Tree;

namespace Land.Core.Markup
{
	public interface IEqualsIgnoreValue
	{
		bool EqualsIgnoreValue(object obj);
	}

	[DataContract(IsReference = true)]
	public abstract class TypedPrioritizedContextElement
	{
		[DataMember]
		public double Priority { get; set; }

		[DataMember]
		public string Type { get; set; }
	}

	[DataContract(IsReference = true)]
	public class HeaderContextElement: TypedPrioritizedContextElement, IEqualsIgnoreValue
	{
		[DataMember]
		public bool ExactMatch { get; set; }

		[DataMember]
		public List<string> Value { get; set; }

		/// Проверка двух контекстов на совпадение всех полей, кроме поля Value
		public bool EqualsIgnoreValue(object obj)
		{
			if (obj is HeaderContextElement elem)
			{
				return ReferenceEquals(this, elem) || Priority == elem.Priority
					&& Type == elem.Type
					&& ExactMatch == elem.ExactMatch;
			}

			return false;
		}

		public override bool Equals(object obj)
		{
			if(obj is HeaderContextElement elem)
			{
				return ReferenceEquals(this, elem) || Priority == elem.Priority 
					&& Type == elem.Type
					&& ExactMatch == elem.ExactMatch
					&& Value.SequenceEqual(elem.Value);
			}

			return false;
		}

		public static bool operator ==(HeaderContextElement a, HeaderContextElement b)
		{
			return a.Equals(b);
		}

		public static bool operator !=(HeaderContextElement a, HeaderContextElement b)
		{
			return !a.Equals(b);
		}

		public override int GetHashCode()
		{
			return base.GetHashCode();
		}

		public static explicit operator HeaderContextElement(Node node)
		{
			return new HeaderContextElement()
			{
				Type = node.Type,
				Value = node.Value,
				Priority = node.Options.Priority.Value,
				ExactMatch = node.Options.ExactMatch
			};
		}
	}

	[DataContract(IsReference = true)]
	public class InnerContextElement: TypedPrioritizedContextElement, IEqualsIgnoreValue
	{
		public const int MAX_TEXT_LENGTH = 200;

		[DataMember]
		public byte[] Hash { get; set; }

		[DataMember]
		public int TextLength { get; set; }

		[DataMember]
		public string Text { get; set; }

		public InnerContextElement(Node node, string fileText)
		{
			Type = node.Type;
			Priority = node.Options.Priority.Value;

			/// Удаляем из текста все пробельные символы
			var text = System.Text.RegularExpressions.Regex.Replace(
				fileText.Substring(node.Anchor.Start.Offset, node.Anchor.Length.Value), "[\n\r\f\t ]", ""
			);

			TextLength = text.Length;

			/// Хэш от строки можем посчитать, только если длина строки
			/// больше заданной константы
			if (text.Length > FuzzyHashing.MIN_TEXT_LENGTH)
				Hash = FuzzyHashing.GetFuzzyHash(text);

			if (text.Length <= MAX_TEXT_LENGTH)
				Text = text;
		}

		public bool EqualsIgnoreValue(object obj)
		{
			if (obj is InnerContextElement elem)
			{
				return ReferenceEquals(this, elem) || Priority == elem.Priority
					&& Type == elem.Type;
			}

			return false;
		}
	}

	[DataContract(IsReference = true)]
	public class AncestorsContextElement
	{
		[DataMember]
		public string Type { get; set; }

		[DataMember]
		public List<HeaderContextElement> HeaderContext { get; set; }

		public override bool Equals(object obj)
		{
			if (obj is AncestorsContextElement elem)
			{
				return ReferenceEquals(this, elem) || Type == elem.Type
					&& HeaderContext.SequenceEqual(elem.HeaderContext);
			}

			return false;
		}

		public static bool operator ==(AncestorsContextElement a, AncestorsContextElement b)
		{
			return a.Equals(b);
		}

		public static bool operator !=(AncestorsContextElement a, AncestorsContextElement b)
		{
			return !a.Equals(b);
		}

		public override int GetHashCode()
		{
			return base.GetHashCode();
		}

		public static explicit operator AncestorsContextElement(Node node)
		{
			return new AncestorsContextElement()
			{
				Type = node.Type,
				HeaderContext = PointContext.GetHeaderContext(node)
			};
		}
	}

	[DataContract(IsReference = true)]
	public class SiblingsContextElement
	{
		[DataMember]
		public string Type { get; set; }

		public static explicit operator SiblingsContextElement(Node node)
		{
			return new SiblingsContextElement()
			{
				Type = node.Type
			};
		}
	}

	[DataContract(IsReference = true)]
	public class PointContext
	{
		[DataMember]
		public string FileName { get; set; }

		[DataMember]
		public string NodeType { get; set; }

		/// <summary>
		/// Контекст заголовка узла, к которому привязана точка разметки
		/// </summary>
		[DataMember]
		public List<HeaderContextElement> HeaderContext { get; set; }

		/// <summary>
		/// Контекст потомков узла, к которому привязана точка разметки
		/// </summary>
		[DataMember]
		public List<InnerContextElement> InnerContext { get; set; }

		/// <summary>
		/// Контекст предков узла, к которому привязана точка разметки
		/// </summary>
		[DataMember]
		public List<AncestorsContextElement> AncestorsContext { get; set; }

		/// <summary>
		/// Контекст уровня, на котором находится узел, к которому привязана точка разметки
		/// </summary>
		[DataMember]
		public List<SiblingsContextElement> SiblingsContext { get; set; }

		public static List<HeaderContextElement> GetHeaderContext(Node node)
		{
			if (node.Value.Count > 0)
			{
				return new List<HeaderContextElement>() { new HeaderContextElement()
				{
					 Type = node.Type,
					 Priority = 1,
					 Value = node.Value
				}};
			}
			else
			{
				var headerContext = new List<HeaderContextElement>();

				var stack = new Stack<Node>(Enumerable.Reverse(node.Children));

				while(stack.Any())
				{
					var current = stack.Pop();

					if (current.Children.Count == 0 && current.Options.Priority > 0)
						headerContext.Add((HeaderContextElement)current);
					else
					{
						if (current.Type == Grammar.CUSTOM_BLOCK_RULE_NAME)
							for (var i = current.Children.Count - 1; i >= 0; --i)
								stack.Push(current.Children[i]);
					}
				}

				return headerContext;
			}
		}

		public static List<AncestorsContextElement> GetAncestorsContext(Node node)
		{
			var context = new List<AncestorsContextElement>();
			var currentNode = node.Parent;

			while(currentNode != null)
			{
				if(currentNode.Symbol != Grammar.CUSTOM_BLOCK_RULE_NAME)
					context.Add((AncestorsContextElement)currentNode);

				currentNode = currentNode.Parent;
			}

			return context;
		}

		public static List<InnerContextElement> GetInnerContext(TargetFileInfo info)
		{
			var innerContext = new List<InnerContextElement>();

			var stack = new Stack<Node>(Enumerable.Reverse(info.TargetNode.Children));

			while (stack.Any())
			{
				var current = stack.Pop();

				if (current.Children.Count > 0)
				{
					if (current.Type != Grammar.CUSTOM_BLOCK_RULE_NAME)
						innerContext.Add(new InnerContextElement(current, info.FileText));
					else
					{
						for (var i = current.Children.Count - 1; i >= 0; --i)
							stack.Push(current.Children[i]);
					}
				}
			}

			return innerContext;
		}

		public static List<SiblingsContextElement> GetSiblingsContext(Node node)
		{
			return node.Parent != null 
				? node.Parent.Children.Select(c => (SiblingsContextElement)c).ToList()
				: new List<SiblingsContextElement>();
		}

		public static PointContext Create(TargetFileInfo info)
		{
			return new PointContext()
			{
				FileName = info.FileName,
				NodeType = info.TargetNode.Type,
				HeaderContext = GetHeaderContext(info.TargetNode),
				AncestorsContext = GetAncestorsContext(info.TargetNode),
				InnerContext = GetInnerContext(info),
				SiblingsContext = GetSiblingsContext(info.TargetNode)
			};
		}
	}

	public class EqualsIgnoreValueComparer : 
		IEqualityComparer<IEqualsIgnoreValue>
	{
		public bool Equals(IEqualsIgnoreValue e1, 
			IEqualsIgnoreValue e2) => e1.EqualsIgnoreValue(e2);

		public int GetHashCode(IEqualsIgnoreValue e) 
			=> e.GetHashCode();
	}
}
