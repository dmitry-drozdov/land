using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Media.Media3D;
using System.Xml.Linq;

namespace Land.GUI
{
	internal class GoFunc
	{
		public string Name { get; set; } = "";
		public List<string> Args { get; set; } = new List<string>();
		public List<int> ArgsDepth { get; set; } = new List<int>();
		public int ArgsCnt { get; set; } = 0;
		public int Return { get; set; } = 0;
		public List<int> ReturnsDepth { get; set; } = new List<int>();
		public string Receiver { get; set; } = "";

		public bool Empty
		{
			get
			{
				return Name == "" && Args.Count == 0 && Return == 0 && ArgsCnt == 0;
			}
		}
	}

	internal class GoStruct
	{
		public string Name { get; set; } = "";
		public List<string> Types { get; set; } = new List<string>();

		public bool Empty
		{
			get
			{
				return Name == "" && Types.Count == 0;
			}
		}
	}

	internal class GoNode
	{
		public string Type { get; set; } = "";
		public string Name { get; set; } = "";
		public List<GoNode> Children { get; set; } = new List<GoNode>();

		/*public bool Empty
		{
			get
			{
				return Name == "" && Type == "" && Children.Count == 0;
			}
		}*/
		public GoNode(string type)
		{
			Type = type;
		}
		public GoNode(string type, string name)
		{
			Type = type;
			Name = name;
		}
	}

	internal class GoBlock
	{
		public int Depth { get; set; } = 0;
		public List<GoBlock> Children { get; set; } = new List<GoBlock>();

		/*public bool Empty
		{
			get
			{
				return Name == "" && Type == "" && Children.Count == 0;
			}
		}*/
		public GoBlock(int depth)
		{
			Depth = depth;
		}
	}

	internal class GoControl
	{
		public string Type { get; set; } = "root";
		public int Depth { get; set; } = 0;
		public List<GoControl> Children { get; set; } = new List<GoControl>();

		/*public bool Empty
		{
			get
			{
				return Name == "" && Type == "" && Children.Count == 0;
			}
		}*/
		public GoControl(string type, int depth)
		{
			Type = type;
			Depth = depth;
		}
	}


}
