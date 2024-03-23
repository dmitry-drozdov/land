using Land.Core;
using Land.Core.Parsing;
using Land.Core.Parsing.Tree;
using Land.Core.Specification;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NoGUI
{
	public struct BatchWorkerArgument
	{
		public string DirectoryPath { get; set; }
		public List<string> Files { get; set; }
		public List<string> TargetEntityTypes { get; set; }
	}
	class Actor
	{
		private Land.Core.Parsing.BaseParser Parser { get; set; }

		private (Node, Durations)? File_Parse(string fileName, string text, bool enableTracing = false)
		{
			return Parser?.Parse(text, enableTracing);
		}

		public void BuildGrammar(string file)
		{
			var messages = new List<Message>();

			Parser = Builder.BuildParser(
			    GrammarType.LR,
			    File.ReadAllText(file),
			    messages
			);

			if (messages.Any(m => m.Type == MessageType.Error))
			{
				Console.WriteLine("Не удалось сгенерировать парсер");
				Environment.Exit(1);
			}
			else
			{
				Console.WriteLine("Парсер успешно сгенерирован");
			}
		}

		public Statistics Do(BatchWorkerArgument argument)
		{
			var statsPerFile = new Dictionary<string, Statistics>();
			var totalStats = new Statistics();
			var precision = argument.Files.Count / 10 + 1;
			for (var counter = 0; counter < argument.Files.Count; ++counter)
			{
				var file = argument.Files[counter];
				try
				{
					Node root = null;
					Durations stats = null;

					(root, stats) = this.File_Parse(file, File.ReadAllText(file)) ?? (null, null);

					if (Parser.Log.Any(l => l.Type == MessageType.Error))
					{
						Console.WriteLine("errors");
					}
					else
					{
						statsPerFile[file] = Parser.Statistics;
						totalStats += Parser.Statistics;

						if (root == null) continue;
					}
				}
				catch (Exception ex)
				{
					Console.WriteLine(ex.ToString());
				}

				if (counter % precision == 0)
				{
					Console.Write($"{(counter * 100) / argument.Files.Count} => ");
				}
			}
			Console.WriteLine();
			return totalStats;
		}
	}
	internal class Program
	{
		static void Main(string[] args)
		{
			var actor = new Actor();
			actor.BuildGrammar("e:\\phd\\my\\land\\LanD Specifications\\sharp\\golang.land");
			var path = "e:\\phd\\test_repos_light\\";

			var files = new List<string>();
			/// Возможна ошибка при доступе к определённым директориям
			try
			{
				files.AddRange(Directory.GetFiles(path, "*.go", SearchOption.AllDirectories));
			}
			catch
			{
				Console.WriteLine($"Ошибка при получении содержимого каталога, возможно, отсутствуют права доступа");
				Environment.Exit(1);
			}
			Console.WriteLine($"Получено {files.Count} файлов");

			var stats = actor.Do(new BatchWorkerArgument()
			{
				DirectoryPath = path,
				Files = files
			});
			Console.WriteLine(stats.ToString());
		}


	}
}
