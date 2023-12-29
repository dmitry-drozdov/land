using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Media.Media3D;

namespace Land.GUI
{
    internal class GoParseResult
    {
        public string Name { get; set; } = "";
        public List<string> Args { get; set; } = new List<string>();
        public int ArgsCnt { get; set; } = 0;
        public int Return { get; set; } = 0;

        public bool Empty
        {
            get
            {
                return Name == "" && Args.Count == 0 && Return == 0 && ArgsCnt == 0;
            }
        }

        public void Reverse()
        {
            Args.Reverse();
        }
    }


}
