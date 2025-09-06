using System;
using System.Collections.Generic;
using System.Linq;
using System.IO;
using System.Diagnostics;
using System.Text;

using Microsoft.CSharp;

using SpecParsing = Land.Core.Specification.Parsing;
using Land.Core.Specification;
using Land.Core.Lexing;
using Land.Core.Parsing;
using Land.Core.Parsing.LL;
using Land.Core.Parsing.LR;
using System.Text.RegularExpressions;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis;
using System.Reflection;

namespace Land.Core
{
	public static class Builder
	{
		private static readonly string TempDir = Path.GetTempPath();

		private static string TempName(string fname) => Path.Combine(TempDir, fname);


		public static Type BuildLexer(Grammar grammar, string lexerName, List<Message> errors = null)
		{
			// 1) Папка для артефактов (удобнее контролировать пути и имена)
			var tmpDir = Path.Combine(Path.GetTempPath(), "antlr", Guid.NewGuid().ToString("N"));
			Directory.CreateDirectory(tmpDir);

			var g4Path = Path.Combine(tmpDir, $"{lexerName}.g4");
			var csPath = Path.Combine(tmpDir, $"{lexerName}.cs");

			// 2) Генерируем .g4
			var tokensForLines = new Dictionary<int, string>();
			var linesCounter = 2;

			using (var grammarOutput = new StreamWriter(g4Path, false, new UTF8Encoding(false)))
			{
				grammarOutput.WriteLine($"lexer grammar {lexerName};");
				grammarOutput.WriteLine();

				foreach (var token in grammar.Tokens.Values.Where(t => t.Name.StartsWith(Grammar.AUTO_TOKEN_PREFIX, StringComparison.Ordinal)))
				{
					grammarOutput.WriteLine($"{token.Name}: {token.Pattern} ;");
					tokensForLines[++linesCounter] = token.Name.StartsWith(Grammar.AUTO_TOKEN_PREFIX, StringComparison.Ordinal) ? token.Pattern : token.Name;
				}

				foreach (var token in grammar.TokenOrder.Where(t => !string.IsNullOrEmpty(grammar.Tokens[t].Pattern)))
				{
					var isFragment = grammar.Options.GetSymbols(ParsingOption.GROUP_NAME, ParsingOption.FRAGMENT).Contains(token);
					var lineStartGuard = grammar.Tokens[token].LineStart
					    ? $" {{this.InputStream.LA(-1 - Text.Length) == 10 || this.InputStream.LA(-1 - Text.Length) == -1}}?"
					    : string.Empty;

					grammarOutput.WriteLine($"{(isFragment ? "fragment " : string.Empty)}{token}: {grammar.Tokens[token].Pattern}{lineStartGuard} ;");
					tokensForLines[++linesCounter] = grammar.Developerify(token);
				}

				grammarOutput.WriteLine(@"WS: [ \n\r\t\u00A0] -> skip ;");

				if (grammar.Options.IsSet(ParsingOption.GROUP_NAME, ParsingOption.IGNOREUNDEFINED))
					grammarOutput.WriteLine(@"UNDEFINED: . -> skip ;");
				else
					grammarOutput.WriteLine(@"UNDEFINED: . ;");
			}

			// 3) Запускаем ANTLR (C# target) и кладём результат в tmpDir
			//   Важно: пусть jar лежит рядом с приложением: ./Resources/antlr-4.7-complete.jar
			var antlrJar = Path.GetFullPath(Path.Combine(AppContext.BaseDirectory, "Resources", "antlr-4.7-complete.jar"));
			var psi = new ProcessStartInfo
			{
				FileName = "java",
				Arguments = $"-jar \"{antlrJar}\" -Dlanguage=CSharp -o \"{tmpDir}\" \"{g4Path}\"",
				CreateNoWindow = true,
				RedirectStandardOutput = true,
				RedirectStandardError = true,
				UseShellExecute = false,
				WorkingDirectory = tmpDir
			};

			string stdOut, stdErr;
			using (var proc = Process.Start(psi))
			{
				stdOut = proc.StandardOutput.ReadToEnd();
				stdErr = proc.StandardError.ReadToEnd();
				proc.WaitForExit();
			}

			var hasAntlrErrors = !string.IsNullOrWhiteSpace(stdErr);

			if (hasAntlrErrors)
			{
				var lines = stdErr.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries)
						  .Select(s => s.Trim());
				foreach (var line in lines)
				{
					try
					{
						// ожидаемый формат: file:line:col: message...
						var parts = line.Split(new[] { ':' }, 5);
						// подменяем автотермы на "человеческие"
						var autoNames = System.Text.RegularExpressions.Regex.Matches(parts[4], $"{Grammar.AUTO_TOKEN_PREFIX}[0-9]+");
						foreach (System.Text.RegularExpressions.Match m in autoNames)
							parts[4] = parts[4].Replace(m.Value, grammar.Developerify(m.Value));

						var errorToken = tokensForLines[int.Parse(parts[2])];
						var loc = grammar.GetLocation(errorToken);
						errors?.Add(Message.Error($"Token {errorToken}: {parts[4]}", loc, "ANTLR Scanner Generator"));
					}
					catch
					{
						errors?.Add(Message.Error(line, null, "ANTLR Scanner Generator"));
					}
				}
				return null;
			}

			// 4) Компилируем сгенерированный {lexerName}.cs через Roslyn
			if (!File.Exists(csPath))
			{
				// некоторые версии ANTLR могут класть файл в рабочую папку по имени грамматики;
				// подстрахуемся поиском
				var fallback = Directory.EnumerateFiles(tmpDir, $"{lexerName}.cs", SearchOption.AllDirectories).FirstOrDefault();
				if (fallback == null)
				{
					errors?.Add(Message.Error($"ANTLR did not produce {lexerName}.cs", null, "ANTLR Scanner Generator"));
					return null;
				}
				csPath = fallback;
			}

			var source = File.ReadAllText(csPath);
			var syntaxTree = CSharpSyntaxTree.ParseText(source);

			// Все «базовые» сборки текущего рантайма .NET 8:
			var tpa = (AppContext.GetData("TRUSTED_PLATFORM_ASSEMBLIES") as string) ?? string.Empty;
			var references = tpa
			    .Split(new[] { Path.PathSeparator }, StringSplitOptions.RemoveEmptyEntries)
			    .Where(File.Exists)
			    .Select(p => MetadataReference.CreateFromFile(p))
			    .ToList();

			// Добавляем ссылку на Antlr4.Runtime.Standard, который у тебя подключён через NuGet
			var antlrAsm = typeof(Antlr4.Runtime.IToken).Assembly;
			references.Add(MetadataReference.CreateFromFile(antlrAsm.Location));

			var compilation = CSharpCompilation.Create(
			    assemblyName: $"{lexerName}_{Guid.NewGuid():N}",
			    syntaxTrees: new[] { syntaxTree },
			    references: references,
			    options: new CSharpCompilationOptions(OutputKind.DynamicallyLinkedLibrary, optimizationLevel: OptimizationLevel.Release));

			using (var ms = new MemoryStream())
			{
				var emitResult = compilation.Emit(ms);
				if (!emitResult.Success)
				{
					foreach (var diag in emitResult.Diagnostics.Where(d => d.Severity == DiagnosticSeverity.Error))
						errors?.Add(Message.Error(diag.ToString(), null, "Roslyn"));
					return null;
				}

				ms.Position = 0;
				var asm = Assembly.Load(ms.ToArray());
				return asm.GetType(lexerName, throwOnError: true);
			}
		}

		public static Grammar BuildGrammar(GrammarType type, string text, List<Message> log)
		{
			var scanner = new SpecParsing.Scanner();
			scanner.SetSource(text, 0);

			var specParser = new SpecParsing.Parser(scanner)
			{
				ConstructedGrammar = new Grammar(type)
			};

			var success = specParser.Parse();

			log.AddRange(specParser.Log);
			log.AddRange(scanner.Log);

			if (!success)
			{
				return null;
			}

			specParser.ConstructedGrammar.FillPairs();

			return specParser.ConstructedGrammar;
		}

		public static BaseParser BuildParser(GrammarType type, string text, List<Message> messages)
		{
			var builtGrammar = BuildGrammar(type, text, messages);

			if (messages.Count(m => m.Type == MessageType.Error) != 0)
				return null;

			builtGrammar.RebuildUserificationCache();

			BaseTable table = null;

			switch (type)
			{
				case GrammarType.LL:
					table = new TableLL1(builtGrammar);
					break;
				case GrammarType.LR:
					table = new TableLR1(builtGrammar);
					break;
			}

			messages.AddRange(table.CheckValidity());

			table.ExportToCsv(TempName("current_table.csv"));

			var lexerType = BuildLexer(builtGrammar, "CurrentLexer", messages);

			if (messages.Count(m => m.Type == MessageType.Error) != 0)
				return null;

			/// Создаём парсер
			BaseParser parser = null;

			switch (type)
			{
				case GrammarType.LL:
					parser = new Parsing.LL.Parser(builtGrammar,
						new AntlrLexerAdapter(
							(Antlr4.Runtime.ICharStream stream) => (Antlr4.Runtime.Lexer)Activator.CreateInstance(lexerType, stream)
						)
					);
					break;
				case GrammarType.LR:
					parser = new Parsing.LR.Parser(builtGrammar,
						new AntlrLexerAdapter(
							(Antlr4.Runtime.ICharStream stream) => (Antlr4.Runtime.Lexer)Activator.CreateInstance(lexerType, stream)
						)
					);
					break;
			}



			/*var adapter = new AntlrLexerAdapter(
				(Antlr4.Runtime.ICharStream stream) => (Antlr4.Runtime.Lexer)Activator.CreateInstance(lexerType, stream)
			);
			var file = "e:\\phd\\large\\large.cs";

			var watch = Stopwatch.StartNew();
			adapter.SetSourceText(File.ReadAllText(file));
			watch.Stop();
			var d1 = watch.ElapsedMilliseconds;

			watch = Stopwatch.StartNew();
			var tokens = adapter.GetAllTokens();
			watch.Stop();
			var d2 = watch.ElapsedMilliseconds;


			MessageBox.Show($"tokens={tokens.Count}, t_read={d1/1000}c., t_allTokens={(float)d2/1000}c.");*/

			return parser;
		}


		/// <summary>
		/// Генерация библиотеки с парсером
		/// </summary>
		/// <param name="type">Тип парсера (LL или LR)</param>
		/// <param name="text">Грамматика разбираемого формата файлов</param>
		/// <param name="namespace">Пространство имён для сгенерированного парсера</param>
		/// <param name="path">Путь к каталогу, в котором необходимо сгенерировать библиотеку</param>
		/// <param name="messages">Лог генерации парсера</param>
		/// <returns>Признак успешности выполнения операции</returns>
		/*public static bool GenerateLibrary(
			GrammarType type,
			string text,
			string @namespace,
			string path,
			string keyPath,
			List<Message> messages
		)
		{
			const string ANTLR_LIBRARY = "Antlr4.Runtime.Standard.dll";

			/// Строим объект грамматики и проверяем, корректно ли прошло построение
			var builtGrammar = BuildGrammar(type, text, messages);

			/// Проверяем, не появились ли ошибки после построения грамматики
			if (messages.Count(m => m.Type == MessageType.Error) != 0)
				return false;

			builtGrammar.RebuildUserificationCache();

			/// Строим таблицу и проверяем, соответствует ли она указанному типу грамматики
			BaseTable table = null;

			switch (type)
			{
				case GrammarType.LL:
					table = new TableLL1(builtGrammar);
					break;
				case GrammarType.LR:
					table = new TableLR1(builtGrammar);
					break;
			}

			messages.AddRange(table.CheckValidity());

			/// Проверяем, не появились ли ошибки после построения таблицы
			if (messages.Count(m => m.Type == MessageType.Error) != 0)
				return false;

			var lexerFileName = TempName($"{@namespace.Replace('.', '_')}_Lexer.cs");
			var parserFileName = TempName($"{@namespace.Replace('.', '_')}_Parser.cs");
			var grammarFileName = TempName($"{@namespace.Replace('.', '_')}_Grammar.cs");
			var nodeGeneratorFileName = TempName($"{@namespace.Replace('.', '_')}_NodeGenerator.cs");

			BuildLexer(builtGrammar, Path.GetFileNameWithoutExtension(lexerFileName), messages);

			/// Проверяем, не появились ли ошибки после генерации исходников лексера
			if (messages.Count(m => m.Type == MessageType.Error) != 0)
				return false;

			File.WriteAllText(grammarFileName, GetGrammarProviderText(builtGrammar, @namespace));
			File.WriteAllText(parserFileName, GetParserProviderText(@namespace));
			File.WriteAllText(nodeGeneratorFileName, GetNodeGeneratorText(builtGrammar, @namespace));

			//if (!String.IsNullOrEmpty(keyPath) && !File.Exists(keyPath))
			//{

			//    // Создаём файл ключа

			//    Process process = new Process();
			//    ProcessStartInfo startInfo = new ProcessStartInfo()
			//    {
			//        FileName = "cmd.exe",
			//        Arguments = $"/C chcp 1251 | \"Resources/sn.exe\" -k \"{keyPath}\"",
			//        CreateNoWindow = true,
			//        RedirectStandardOutput = true,
			//        UseShellExecute = false
			//    };
			//    process.StartInfo = startInfo;
			//    process.Start();

			//    process.WaitForExit();
			//}

			/// Компилируем библиотеку
			var codeProvider = new CSharpCodeProvider();
			var compilerParams = new System.CodeDom.Compiler.CompilerParameters();

			compilerParams.GenerateInMemory = false;
			compilerParams.OutputAssembly = Path.Combine(path, $"{@namespace}.dll");
			compilerParams.ReferencedAssemblies.Add(ANTLR_LIBRARY);
			compilerParams.ReferencedAssemblies.Add("Land.Core.dll");
			compilerParams.ReferencedAssemblies.Add("System.dll");
			compilerParams.ReferencedAssemblies.Add("System.Core.dll");
			compilerParams.ReferencedAssemblies.Add("mscorlib.dll");

			if (!String.IsNullOrEmpty(keyPath))
				compilerParams.CompilerOptions = $"/keyfile:\"{keyPath}\"";

			var compilationResult = codeProvider.CompileAssemblyFromFile(compilerParams, lexerFileName, grammarFileName, parserFileName, nodeGeneratorFileName);

			if (compilationResult.Errors.Count == 0)
			{
				File.Copy(ANTLR_LIBRARY, Path.Combine(path, ANTLR_LIBRARY), true);

				return true;
			}
			else
			{
				foreach (System.CodeDom.Compiler.CompilerError error in compilationResult.Errors)
				{
					if (error.IsWarning)
					{
						messages.Add(Message.Warning(
							$"Предупреждение: {error.FileName}; ({error.Line}, {error.Column}); {error.ErrorText}",
							null,
							"C# Compiler"
						));
					}
					else
					{
						messages.Add(Message.Error(
							$"Ошибка: {error.FileName}; ({error.Line}, {error.Column}); {error.ErrorText}",
							null,
							"C# Compiler"
						));
					}
				}

				return messages.All(m => m.Type != MessageType.Error);
			}
		}*/

		private static string GetGrammarProviderText(Grammar grammar, string @namespace)
		{
			return
@"
using Land.Core.Specification;
using System.Collections.Generic;

namespace " + @namespace + @"
{
	public static class GrammarProvider
	{
		public static Grammar GetGrammar()
		{
" + String.Join(Environment.NewLine, grammar.ConstructionLog.Select(rec => $"\t\t\t{rec}")) + @"
			grammar.ForceValid();
			return grammar;
		}
	}
}";
		}

		private static string GetParserProviderText(string @namespace)
		{
			return
@"
using System;
using System.Reflection;
using Land.Core.Specification;
using Land.Core.Lexing;
using Land.Core.Parsing;
using Land.Core.Parsing.Tree;

namespace " + @namespace + @"
{
	public static class ParserProvider
	{
		public static BaseParser GetParser(bool buildTypedTree = true)
		{
			var grammar = GrammarProvider.GetGrammar();
			var lexerType = Assembly.GetExecutingAssembly().GetType(""" + @namespace.Replace('.', '_') + @"_Lexer"");

			BaseParser parser = null;

			switch (grammar.Type)
			{
				case GrammarType.LL:
					parser = new Land.Core.Parsing.LL.Parser(grammar,
						new AntlrLexerAdapter(
							(Antlr4.Runtime.ICharStream stream) => (Antlr4.Runtime.Lexer)Activator.CreateInstance(lexerType, stream)
						),
						buildTypedTree ? new NodeGenerator(grammar) : null,
						buildTypedTree ? new NodeRetypingVisitor(grammar) : null
					);
					break;
				case GrammarType.LR:
					parser = new Land.Core.Parsing.LR.Parser(grammar,
						new AntlrLexerAdapter(
							(Antlr4.Runtime.ICharStream stream) => (Antlr4.Runtime.Lexer)Activator.CreateInstance(lexerType, stream)
						),
						buildTypedTree ? new NodeGenerator(grammar) : null,
						buildTypedTree ? new NodeRetypingVisitor(grammar) : null
					);
					break;
			}

			return parser;
		}
	}
}";
		}

		private static string GetNodeGeneratorText(Grammar grammar, string @namespace)
		{
			var nodeClassesSource = new StringBuilder();

			nodeClassesSource.AppendLine("using System;");
			nodeClassesSource.AppendLine("using System.Collections.Generic;");
			nodeClassesSource.AppendLine("using System.Linq;");
			nodeClassesSource.AppendLine("using System.Reflection;");

			nodeClassesSource.AppendLine("using Land.Core.Specification;");
			nodeClassesSource.AppendLine("using Land.Core.Parsing.Tree;");

			nodeClassesSource.AppendLine($"namespace {@namespace} {{");

			nodeClassesSource.AppendLine(@"
	public class NodeGenerator : BaseNodeGenerator 
	{
		public const string BASE_RULE_TYPE = ""RuleNode"";
		public const string BASE_TOKEN_TYPE = ""TokenNode"";

		public NodeGenerator(Grammar grammar): base(grammar)
		{
			Cache[BASE_RULE_TYPE] = Assembly.GetExecutingAssembly().GetType(""" + @namespace + @"."" + BASE_RULE_TYPE)
				.GetConstructor(new Type[] { typeof(string), typeof(SymbolOptionsManager), typeof(SymbolArguments) }) ;
			Cache[BASE_TOKEN_TYPE] = Assembly.GetExecutingAssembly().GetType(""" + @namespace + @"."" + BASE_TOKEN_TYPE)
				.GetConstructor(new Type[] { typeof(string), typeof(SymbolOptionsManager), typeof(SymbolArguments) }) ;

			foreach (var smb in grammar.Rules.Keys)
			{
				if(smb.StartsWith(Grammar.AUTO_RULE_PREFIX, StringComparison.Ordinal))
					Cache[smb] = Cache[BASE_RULE_TYPE];
				else
				{
					var type = Assembly.GetExecutingAssembly().GetType(""" + @namespace + @"."" + smb + ""_node"");
					Cache[smb] = type != null ? type.GetConstructor(new Type[] { typeof(string), typeof(SymbolOptionsManager), typeof(SymbolArguments) }) : Cache[BASE_RULE_TYPE];
				}
			}

			foreach (var smb in grammar.Tokens.Keys)
			{
				if(smb.StartsWith(Grammar.AUTO_TOKEN_PREFIX, StringComparison.Ordinal))
					Cache[smb] = Cache[BASE_TOKEN_TYPE];
				else
				{
					var type = Assembly.GetExecutingAssembly().GetType(""" + @namespace + @"."" + smb + ""_node"");
					Cache[smb] = type != null ? type.GetConstructor(new Type[] { typeof(string), typeof(SymbolOptionsManager), typeof(SymbolArguments) }) : Cache[BASE_TOKEN_TYPE];
				}
			}
		}
	}");

			nodeClassesSource.AppendLine(@"
	public class NodeRetypingVisitor : BaseNodeRetypingVisitor
	{
		private Dictionary<string, ConstructorInfo> Cache { get; set; }

		public NodeRetypingVisitor(Grammar grammar): base(grammar)
		{
			Cache = new Dictionary<string, ConstructorInfo>();

			foreach (var kvp in grammar.Aliases)
				foreach (var alias in kvp.Value)
				{
					var type = Assembly.GetExecutingAssembly().GetType(""" + @namespace + @"."" + alias + ""_node"");

					if(type != null)
						Cache[alias] = type.GetConstructor(new Type[] { typeof(Node) });
				}
		}

		public override void Visit(Node node)
		{
			for (var i = 0; i < node.Children.Count; ++i)
				node.Children[i].Accept(this);

			if(!String.IsNullOrEmpty(node.Alias) && Cache.ContainsKey(node.Alias))
			{
				var newNode = (Node)Cache[node.Alias].Invoke(new object[] { node });

				if(node.Parent != null)
				{
					var idx = node.Parent.Children.IndexOf(node);
					node.Parent.Children.RemoveAt(idx);
					node.Parent.Children.Insert(idx, newNode);
				}
				else
					Root = newNode;
				
				foreach(var child in newNode.Children)
					child.Parent = newNode;

				node = newNode;
			}
		}
	}
	
	[Serializable]
	public class TypedNode: Node
	{
		public IEnumerable<TypedNode> TypedChildren 
		{ 
			get 
			{
				return Children.Select(e => (TypedNode)e);
			}
		}

		public TypedNode(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public TypedNode(Node node): base(node) {}

		public virtual void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}

	[Serializable]
	public class RuleNode: TypedNode
	{
		public RuleNode(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public RuleNode(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}
	
	[Serializable]
	public class TokenNode: TypedNode
	{
		public TokenNode(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public TokenNode(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}

	[Serializable]
	public class " + Grammar.CUSTOM_BLOCK_RULE_NAME + @"_node : RuleNode
	{
		public " + Grammar.CUSTOM_BLOCK_RULE_NAME + @"_node(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public " + Grammar.CUSTOM_BLOCK_RULE_NAME + @"_node(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}");


			foreach (var name in grammar.Rules.Keys.Where(key => !key.StartsWith(Grammar.AUTO_RULE_PREFIX, StringComparison.Ordinal)))
				nodeClassesSource.AppendLine(@"
	[Serializable]
	public class " + name + @"_node : RuleNode 
	{
		public " + name + @"_node(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public " + name + @"_node(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}");

			foreach (var kvp in grammar.Aliases)
				foreach (var alias in kvp.Value)
					nodeClassesSource.AppendLine(@"
	[Serializable]
	public class " + alias + @"_node : " + kvp.Key + @"_node 
	{
		public " + alias + @"_node(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public " + alias + @"_node(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}");

			foreach (var name in grammar.Tokens.Keys.Where(key => !key.StartsWith(Grammar.AUTO_TOKEN_PREFIX, StringComparison.Ordinal)))
				nodeClassesSource.AppendLine(@"
	[Serializable]
	public class " + name + @"_node : TokenNode 
	{
		public " + name + @"_node(string symbol, SymbolOptionsManager opts = null, SymbolArguments args = null): base(symbol, opts, args) {}
		public " + name + @"_node(Node node): base(node) {}

		public override void Accept(BaseTypedTreeVisitor visitor)
		{
			visitor.Visit(this);
		}
	}");

			nodeClassesSource.AppendLine(@"
	public class BaseTypedTreeVisitor 
	{
		public virtual void Visit(TypedNode node)
		{
			foreach (var child in node.TypedChildren)
				child.Accept(this);
		}

		public virtual void Visit(RuleNode node)
		{
			foreach (var child in node.TypedChildren)
				child.Accept(this);
		}

		public virtual void Visit(TokenNode node) {}");
			foreach (var name in grammar.Rules.Keys.Where(key => !key.StartsWith(Grammar.AUTO_RULE_PREFIX, StringComparison.Ordinal)))
				nodeClassesSource.AppendLine(@"
		public virtual void Visit(" + name + @"_node node)
		{
			foreach (var child in node.TypedChildren)
				child.Accept(this);
		}");

			foreach (var kvp in grammar.Aliases)
				foreach (var alias in kvp.Value)
					nodeClassesSource.AppendLine(@"
		public virtual void Visit(" + alias + @"_node node)
		{
			foreach (var child in node.TypedChildren)
				child.Accept(this);
		}");

			foreach (var name in grammar.Tokens.Keys.Where(key => !key.StartsWith(Grammar.AUTO_TOKEN_PREFIX, StringComparison.Ordinal)))
				nodeClassesSource.AppendLine(@"
		public virtual void Visit(" + name + @"_node node) {}");


			nodeClassesSource.AppendLine(@"
	}");
			nodeClassesSource.AppendLine("}");

			return nodeClassesSource.ToString();
		}
	}
}
