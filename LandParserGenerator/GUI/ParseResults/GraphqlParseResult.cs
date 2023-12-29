using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using System.Windows.Media.Media3D;

namespace Land.GUI
{
    internal class GraphqlParseResult
    {
        public List<GraphqlInput> Inputs { get; set; } = new List<GraphqlInput>();
        public List<GraphqlType> Types { get; set; } = new List<GraphqlType>();
        public List<GraphqlFunc> Funcs { get; set; } = new List<GraphqlFunc>();

        [JsonIgnore]
        public bool Empty
        {
            get
            {
                return Inputs.Count == 0 && Types.Count == 0 && Funcs.Count == 0;
            }
        }
    }

    internal class GraphqlInput
    {
        public string Name { get; set; } = "";
        public List<GraphqlDef> Defs { get; set; } = new List<GraphqlDef>();
        public GraphqlInput(GraphqlType v)
        {
            Name = v.Name;
            Defs = v.Defs;
        }
    }

    internal class GraphqlType
    {
        public string Name { get; set; } = "";
        public List<GraphqlDef> Defs { get; set; } = new List<GraphqlDef>();
    }

    internal class GraphqlFunc
    {
        public string Name { get; set; } = "";
        public List<GraphqlDef> Args { get; set; } = new List<GraphqlDef>();
        public string Return { get; set; } = "";
    }

    internal class GraphqlDef
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
    }
}
