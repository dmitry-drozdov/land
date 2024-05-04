using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Media.Media3D;

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

}
