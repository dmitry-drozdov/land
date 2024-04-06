using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Land.Core.Specification;

namespace Land.Core.Parsing.LR
{
	public struct Item
	{
		public HashSet<Marker> Markers { get; set; }
		public HashSet<Marker> AnyProvokedMarkers { get; set; }
		public Entry AnyEntry { get; set; }
	}
}
