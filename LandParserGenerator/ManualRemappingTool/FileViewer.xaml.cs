﻿using ICSharpCode.AvalonEdit;
using Land.Control;
using Land.Core;
using Land.Core.Parsing.Tree;
using Land.Markup.CoreExtension;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;

namespace ManualRemappingTool
{
	/// <summary>
	/// Логика взаимодействия для FileViewer.xaml
	/// </summary>
	public partial class FileViewer : UserControl
	{
		private SegmentsBackgroundRenderer SegmentColorizer { get; set; }

		/// <summary>
		/// Панель поиска по тексту
		/// </summary>
		private EditorSearchHandler QuickSearch { get; set; }

		/// <summary>
		/// Корень дерева, соответствующего открытому тексту
		/// </summary>
		private Node TreeRoot { get; set; }

		/// <summary>
		/// Упорядоченный по смещению список сущностей, к которым возможна привязка
		/// </summary>
		public List<Node> EntitiesAvailable { get; private set; }

		/// <summary>
		/// Индекс текущей выделенной сущности в списке Entities
		/// </summary>
		private int SelectedEntityIdx { get; set; } = -1;

		/// <summary>
		/// Список файлов, содержащихся в каталоге, 
		/// с которым ведётся работа в рамках данного экземпляра просмотрщика
		/// </summary>
		private List<string> WorkingDirectoryFiles { get; set; }

		/// <summary>
		/// Индекс текущего открытого файла в общем списке файлов каталога
		/// </summary>
		private int CurrentFileIndex { get; set; } = -1;

		/// <summary>
		/// Менеджер, предоставляющий парсеры для разбора
		/// </summary>
		public ParserManager Parsers { get; set; }

		public HashSet<string> WorkingExtensions
		{
			get { return _workingExtensions; }

			set
			{
				_workingExtensions = value;

				if (!String.IsNullOrEmpty(WorkingDirectory)
					&& WorkingExtensions != null && WorkingExtensions.Count > 0)
				{
					WorkingDirectoryFiles = Directory
						.GetFiles(WorkingDirectory, "*", SearchOption.AllDirectories)
						.Where(elem => WorkingExtensions.Contains(Path.GetExtension(elem)))
						.OrderBy(elem => elem)
						.ToList();
				}
			}
		}

		private HashSet<string> _workingExtensions;

		public int? EntityStartOffset => (FileEntitiesList.SelectedItem as ExistingConcernPointCandidate)
			?.Node?.Location?.Start.Offset ?? null;

		public string EntityType => (FileEntitiesList.SelectedItem as ExistingConcernPointCandidate)
			?.Node?.Type ?? null;


		#region Dependency properties

		public static readonly DependencyProperty WorkingDirectoryProperty = DependencyProperty.Register(
			"WorkingDirectory",
			typeof(string),
			typeof(FileViewer),
			new FrameworkPropertyMetadata(null, FrameworkPropertyMetadataOptions.BindsTwoWayByDefault)
		);

		public string WorkingDirectory
		{
			get => (string)GetValue(WorkingDirectoryProperty);

			set
			{
				SetValue(WorkingDirectoryProperty, value);

				if (!String.IsNullOrEmpty(WorkingDirectory)
					&& WorkingExtensions != null && WorkingExtensions.Count > 0)
				{
					WorkingDirectoryFiles = Directory
						.GetFiles(WorkingDirectory, "*", SearchOption.AllDirectories)
						.Where(elem => WorkingExtensions.Contains(Path.GetExtension(elem)))
						.OrderBy(elem => elem)
						.ToList();
				}
			}
		}

		public static readonly DependencyProperty LabelTextProperty = DependencyProperty.Register(
			"LabelText",
			typeof(string),
			typeof(FileViewer),
			new FrameworkPropertyMetadata(null, FrameworkPropertyMetadataOptions.BindsTwoWayByDefault)
		);

		public string LabelText
		{
			get => (string)GetValue(LabelTextProperty);
			set { SetValue(LabelTextProperty, value); }
		}

		public static readonly DependencyProperty FilePathProperty = DependencyProperty.Register(
			"FilePath",
			typeof(string),
			typeof(FileViewer),
			new FrameworkPropertyMetadata(null)
		);

		public string FilePath
		{
			get => (string)GetValue(FilePathProperty);

			private set 
			{
				SetValue(FilePathProperty, 
					GetRelativePath(value, WorkingDirectory)); 
			}
		}

		public static readonly DependencyProperty AreNextPrevEnabledProperty = DependencyProperty.Register(
			"AreNextPrevEnabled",
			typeof(bool),
			typeof(FileViewer),
			new FrameworkPropertyMetadata(true, FrameworkPropertyMetadataOptions.BindsTwoWayByDefault)
		);

		public bool AreNextPrevEnabled
		{
			get => (bool)GetValue(AreNextPrevEnabledProperty);
			set { SetValue(AreNextPrevEnabledProperty, value); }
		}

		#endregion

		#region Events

		public event EventHandler<string> FileOpened;

		public event EventHandler<string> MessageSent;

		#endregion

		public FileViewer()
		{
			InitializeComponent();

			FontSize = 14;
			QuickSearch = new EditorSearchHandler(FileEditor.TextArea);
			SegmentColorizer = new SegmentsBackgroundRenderer(FileEditor.TextArea);
		}

		private void OpenFileButton_Click(object sender, RoutedEventArgs e)
		{
			if(String.IsNullOrEmpty(WorkingDirectory))
			{
				MessageSent?.Invoke(this,
					$"Необходимо указать рабочий каталог для файлов {LabelText}");

				return;
			}

			var openFileDialog = new OpenFileDialog();
			openFileDialog.InitialDirectory = WorkingDirectory;

			if (openFileDialog.ShowDialog() == true 
				&& openFileDialog.FileName.StartsWith(WorkingDirectory))
			{
				OpenFile(openFileDialog.FileName);
				FileOpened?.Invoke(this, FilePath);
			}
		}

		private void OpenPrevFileButton_Click(object sender, RoutedEventArgs e)
		{
			OpenPrevFile();
		}

		private void OpenNextFileButton_Click(object sender, RoutedEventArgs e)
		{
			OpenNextFile();
		}

		private void FileEditor_PreviewMouseUp(object sender, MouseButtonEventArgs e)
		{
			if(Keyboard.Modifiers.HasFlag(ModifierKeys.Control))
			{
				FillEntitiesList(FileEditor.CaretOffset);
			}
		}

		private void FileEntitiesList_SelectionChanged(object sender, SelectionChangedEventArgs e)
		{
			var selected = (ExistingConcernPointCandidate)FileEntitiesList.SelectedItem;

			SelectedEntityIdx = EntitiesAvailable.IndexOf(selected?.Node);

			SegmentColorizer.ResetSegments();

			if (selected != null 
				&& FileEntitiesList.SelectedIndex != FileEntitiesList.Items.Count - 1)
			{
				SegmentColorizer.SetSegments(
					new List<SegmentLocation> { selected.Node.Location }, 
					Color.FromRgb(170, 210, 170)
				);
				FileEditor.ScrollToLine(
					FileEditor.Document.GetLineByOffset(selected.Node.Location.Start.Offset).LineNumber
				);
			}
		}

		private void FileEditor_KeyDown(object sender, KeyEventArgs e)
		{
			if (Keyboard.Modifiers == ModifierKeys.None)
			{
				if (Keyboard.IsKeyDown(Key.S))
				{
					MoveToNextAvailableEntity();
					e.Handled = true;
				}
				else if (Keyboard.IsKeyDown(Key.W))
				{
					MoveToPrevAvailableEntity();
					e.Handled = true;
				}
				else if (Keyboard.IsKeyDown(Key.D))
				{
					OpenNextFile();
					e.Handled = true;
				}
				else if (Keyboard.IsKeyDown(Key.A))
				{
					OpenPrevFile();
					e.Handled = true;
				}
			}
		}

		#region API

		public void OpenFile(string filePath)
		{
			FilePath = filePath;

			using (var stream = new StreamReader(filePath, GetEncoding(filePath)))
			{
				FileEditor.Text = stream.ReadToEnd();
				FileEditor.Encoding = stream.CurrentEncoding;
				FileEditor.ScrollToLine(0);

				SegmentColorizer.ResetSegments();

				FileEditor.SyntaxHighlighting = ICSharpCode.AvalonEdit.Highlighting.HighlightingManager
					.Instance.GetDefinitionByExtension(Path.GetExtension(filePath));
			}

			TreeRoot = Parse(filePath, FileEditor.Text);

			var visitor = new LandExplorerVisitor();
			TreeRoot.Accept(visitor);
			EntitiesAvailable = visitor.Land;
		}

		public void FillEntitiesList(int offset)
		{
			var candidates = GetEntities(TreeRoot, new PointLocation(offset));

			candidates.Add(new ExistingConcernPointCandidate() { ViewHeader = "[сбросить выделение]" });

			FileEntitiesList.ItemsSource = candidates;
			FileEntitiesList.SelectedIndex = 0;
		}

		public void OpenPrevFile()
		{
			if (WorkingDirectoryFiles.Count > 0)
			{
				ShiftToFileCore(
					(WorkingDirectoryFiles.Count + (CurrentFileIndex - 1)) % WorkingDirectoryFiles.Count
				);
			}
		}

		public void OpenNextFile()
		{
			if (WorkingDirectoryFiles.Count > 0)
			{
				ShiftToFileCore(
					(CurrentFileIndex + 1) % WorkingDirectoryFiles.Count
				);
			}
		}

		public void MoveToNextAvailableEntity()
		{
			if (EntitiesAvailable.Count > 0)
			{
				ShiftToEntityCore(
					(SelectedEntityIdx + 1) % EntitiesAvailable.Count
				);
			}
		}

		public void MoveToPrevAvailableEntity()
		{
			if (EntitiesAvailable.Count > 0)
			{
				ShiftToEntityCore(
					(EntitiesAvailable.Count + (SelectedEntityIdx - 1)) % EntitiesAvailable.Count
				);
			}
		}

		#endregion

		#region Cores

		private void ShiftToFileCore(int index)
		{
			CurrentFileIndex = index;

			OpenFile(WorkingDirectoryFiles[CurrentFileIndex]);
			FileOpened?.Invoke(this, FilePath);
		}

		private void ShiftToEntityCore(int index)
		{
			SelectedEntityIdx = index;

			var candidates = GetEntities(
				EntitiesAvailable[SelectedEntityIdx],
				new PointLocation(EntitiesAvailable[SelectedEntityIdx].Location.Start.Offset),
				false
			);
			candidates.Add(new ExistingConcernPointCandidate() { ViewHeader = "[сбросить выделение]" });

			FileEntitiesList.ItemsSource = candidates;
			FileEntitiesList.SelectedIndex = 0;
		}

		#endregion

		#region Helpers

		private Encoding GetEncoding(string filename)
		{
			using (FileStream fs = File.OpenRead(filename))
			{
				Ude.CharsetDetector cdet = new Ude.CharsetDetector();
				cdet.Feed(fs);
				cdet.DataEnd();
				if (cdet.Charset != null)
				{
					return Encoding.GetEncoding(cdet.Charset);
				}
				else
				{
					return Encoding.Default;
				}
			}
		}

		private Node Parse(string fileName, string text)
		{
			if (!String.IsNullOrEmpty(fileName))
			{
				var extension = Path.GetExtension(fileName);

				if (Parsers[extension] != null)
				{
					Node root = null;

					root = Parsers[extension].Parse(text);
					var success = Parsers[extension].Log.All(l => l.Type != MessageType.Error);

					return success ? root : null;
				}
			}

			return null;
		}

		private static string GetRelativePath(string filePath, string directoryPath)
		{
			var directoryUri = new Uri(directoryPath + "/");

			return Uri.UnescapeDataString(
				directoryUri.MakeRelativeUri(new Uri(filePath)).ToString()
			);
		}

		private List<ConcernPointCandidate> GetEntities(Node root, PointLocation point, bool exploreInDepth = true)
		{
			var pseudoSegment = new SegmentLocation
			{
				End = point,
				Start = point
			};

			var pointCandidates = new LinkedList<Node>();
			var currentNode = root;

			/// В качестве кандидатов на роль помечаемого участка рассматриваем узлы от корня,
			/// содержащие текущую позицию каретки
			while (currentNode != null)
			{
				if (currentNode.Options.IsSet(MarkupOption.GROUP_NAME, MarkupOption.LAND))
					pointCandidates.AddFirst(currentNode);

				currentNode = currentNode.Children
					.Where(c => c.Location != null && c.Location.Includes(pseudoSegment))
					.FirstOrDefault();
			}

			return pointCandidates
				.Where(c=>EntitiesAvailable.Contains(c))
				.Select(c => (ConcernPointCandidate)new ExistingConcernPointCandidate(c))
				.ToList();
		}

		#endregion
	}
}
