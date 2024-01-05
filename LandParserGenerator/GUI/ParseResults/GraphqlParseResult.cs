using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
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

        private Dictionary<string, int> FindType = new Dictionary<string, int>();

        [JsonIgnore]
        public bool Empty
        {
            get
            {
                return Inputs.Count == 0 && Types.Count == 0 && Funcs.Count == 0;
            }
        }

        public void AddInput(GraphqlInput input) => Inputs.Add(input);
        public void AddFunc(GraphqlFunc func) => Funcs.Add(func);
        public void AddType(GraphqlType type)
        {
            if (FindType.TryGetValue(type.Name, out int idx))
            {
                var existType = Types[idx];
                if (existType == type)
                {
                    //Console.WriteLine("identical duplicated");
                    return;
                }
                //Console.WriteLine("extension");

                existType.Defs.AddRange(type.Defs);
                return;
            }
            FindType.Add(type.Name, Types.Count);
            Types.Add(type);
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

    internal class GraphqlType : IEquatable<GraphqlType>
    {
        public string Name { get; set; } = "";
        public List<GraphqlDef> Defs { get; set; } = new List<GraphqlDef>();

        public static bool operator ==(GraphqlType obj1, GraphqlType obj2)
        {
            if (ReferenceEquals(obj1, obj2))
                return true;
            if (obj1 is null || obj2 is null)
                return false;
            return obj1.Equals(obj2);
        }
        public static bool operator !=(GraphqlType obj1, GraphqlType obj2) => !(obj1 == obj2);
        public bool Equals(GraphqlType other)
        {
            if (ReferenceEquals(other, null))
                return false;
            if (ReferenceEquals(this, other))
                return true;
            return Name.Equals(other.Name) && Defs.SequenceEqual(other.Defs);
        }
        public override bool Equals(object obj) => Equals(obj as GraphqlType);
    }

    internal class GraphqlFunc
    {
        public string Parent { get; set; } = "";
        public string Name { get; set; } = "";
        public List<GraphqlDef> Args { get; set; } = new List<GraphqlDef>();
        public string Return { get; set; } = "";
    }

    internal class GraphqlDef : IEquatable<GraphqlDef>
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";

        public static bool operator ==(GraphqlDef obj1, GraphqlDef obj2)
        {
            if (ReferenceEquals(obj1, obj2))
                return true;
            if (obj1 is null || obj2 is null)
                return false;
            return obj1.Equals(obj2);
        }
        public static bool operator !=(GraphqlDef obj1, GraphqlDef obj2) => !(obj1 == obj2);
        public bool Equals(GraphqlDef other)
        {
            if (ReferenceEquals(other, null))
                return false;
            if (ReferenceEquals(this, other))
                return true;
            return Name.Equals(other.Name) && Type.Equals(other.Type);
        }
        public override bool Equals(object obj) => Equals(obj as GraphqlDef);
    }
}
