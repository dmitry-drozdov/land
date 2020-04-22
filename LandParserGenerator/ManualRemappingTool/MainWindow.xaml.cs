﻿using Land.Control;
using Land.Core;
using ManualRemappingTool.Properties;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.IO;
using System.Runtime.Serialization;
using System.Windows;
using System.Windows.Input;

namespace ManualRemappingTool
{
	/// <summary>
	/// Логика взаимодействия для MainWindow.xaml
	/// </summary>
	public partial class MainWindow : Window
	{
		#region Consts

		private const int MIN_FONT_SIZE = 8;
		private const int MAX_FONT_SIZE = 40;

		private static readonly string APP_DATA_DIRECTORY =
			Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData) + @"\LanD Control";
		private static readonly string CACHE_DIRECTORY =
			Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData) + @"\LanD Control\Cache";
		public static readonly string SETTINGS_FILE_NAME = "LandExplorerSettings.xml";

		public static string SETTINGS_DEFAULT_PATH =>
			System.IO.Path.Combine(APP_DATA_DIRECTORY, SETTINGS_FILE_NAME);

		#endregion

		private object FileOpeningLock { get; set; } = new object();
		private ParserManager Parsers { get; set; } = new ParserManager();
		private Dataset Dataset { get; set; }
		
		public MainWindow()
		{
			InitializeComponent();

			SourceFileView.Parsers = Parsers;
			TargetFileView.Parsers = Parsers;

			Parsers.Load(LoadSettings(SETTINGS_DEFAULT_PATH), CACHE_DIRECTORY, new List<Message>());
		}

		private void MainWindow_ContentRendered(object sender, EventArgs e)
		{
			StartWindowInteraction();
		}

		private void Control_PreviewMouseWheel(object sender, MouseWheelEventArgs e)
		{
			var controlSender = sender as System.Windows.Controls.Control;

			if (Keyboard.PrimaryDevice.Modifiers == ModifierKeys.Control)
			{
				e.Handled = true;

				if (e.Delta > 0 && controlSender.FontSize < MAX_FONT_SIZE)
					++controlSender.FontSize;
				else if (controlSender.FontSize > MIN_FONT_SIZE)
					--controlSender.FontSize;
			}
		}

		private void NewDatasetButton_Click(object sender, RoutedEventArgs e)
		{
			if (Dataset.Records.Count > 0)
			{
				switch (MessageBox.Show(
						"Сохранить текущий датасет?",
						String.Empty,
						MessageBoxButton.YesNoCancel,
						MessageBoxImage.Question))
				{
					case MessageBoxResult.Yes:
						SaveDatasetButton_Click(sender, e);
						break;
					case MessageBoxResult.No:
						break;
					case MessageBoxResult.Cancel:
						return;
				}
			}

			Dataset.New();
		}

		private void LoadDatasetButton_Click(object sender, RoutedEventArgs e)
		{
			if (Dataset.Records.Count > 0) {
				switch (MessageBox.Show(
						"Сохранить текущий датасет?",
						String.Empty,
						MessageBoxButton.YesNoCancel,
						MessageBoxImage.Question))
				{
					case MessageBoxResult.Yes:
						SaveDatasetButton_Click(sender, e);
						break;
					case MessageBoxResult.No:
						break;
					case MessageBoxResult.Cancel:
						return;
				}
			}

			StartWindowInteraction();
		}

		private void SaveDatasetButton_Click(object sender, RoutedEventArgs e)
		{
			if(!String.IsNullOrEmpty(Dataset.SavingPath))
			{
				Dataset.Save();
			}
			else
			{
				SaveDatasetAsButton_Click(sender, e);
			}
		}

		private void SaveDatasetAsButton_Click(object sender, RoutedEventArgs e)
		{
			var saveFileDialog = new SaveFileDialog()
			{
				AddExtension = true,
				DefaultExt = "ds.txt",
				Filter = "Текстовые файлы (*.ds.txt)|*.ds.txt|Все файлы (*.*)|*.*"
			};

			if (saveFileDialog.ShowDialog() == true)
			{
				Dataset.SavingPath = saveFileDialog.FileName;
				SaveDatasetButton_Click(sender, e);
			}
		}

		private void AddToDatasetButton_Click(object sender, RoutedEventArgs e)
		{
			Dataset.Add(
					SourceFileView.FilePath,
					TargetFileView.FilePath,
					SourceFileView.EntityStartLine.Value,
					TargetFileView.EntityStartLine.Value,
					SourceFileView.EntityType
				);
		}

		private void RemoveFromDatasetButton_Click(object sender, RoutedEventArgs e)
		{
			Dataset.Remove(
					SourceFileView.FilePath,
					TargetFileView.FilePath,
					SourceFileView.EntityStartLine.Value,
					TargetFileView.EntityStartLine.Value,
					SourceFileView.EntityType
				);
		}

		private void HaveDoubtsButton_Click(object sender, RoutedEventArgs e)
		{
			Dataset.Add(
					SourceFileView.FilePath,
					TargetFileView.FilePath,
					SourceFileView.EntityStartLine.Value,
					TargetFileView.EntityStartLine.Value,
					SourceFileView.EntityType,
					true
				);
		}

		private void SourceFileView_FileOpened(object sender, string e)
		{
			lock(FileOpeningLock)
			{
				TargetFileView.OpenFile(e);
			}
		}

		private void TargetFileView_FileOpened(object sender, string e)
		{
			lock (FileOpeningLock)
			{
				SourceFileView.OpenFile(e);
			}
		}

		private void Control_MessageSent(object sender, string e)
		{
			this.Title = e;
		}

		#region Helpers

		private LandExplorerSettings LoadSettings(string path)
		{
			if (File.Exists(path))
			{
				var serializer = new DataContractSerializer(
					typeof(LandExplorerSettings), new Type[] { typeof(ParserSettingsItem) }
				);

				using (FileStream fs = new FileStream(path, FileMode.Open))
				{
					return (LandExplorerSettings)serializer.ReadObject(fs);
				}
			}
			else
			{
				return null;
			}
		}

		private StartWindow CreateStartWindow()
		{
			var startWindow = new StartWindow();

			startWindow.Owner = this;

			startWindow.Dataset = this.Dataset ?? new Dataset();
			startWindow.Dataset.New();

			return startWindow;
		}

		private void StartWindowInteraction()
		{
			var startWindow = CreateStartWindow();

			if (startWindow.ShowDialog() ?? false)
			{
				Dataset = startWindow.Dataset;

				SourceFileView.WorkingDirectory = Dataset.SourceDirectoryPath;
				TargetFileView.WorkingDirectory = Dataset.TargetDirectoryPath;

				SourceFileView.WorkingExtensions = TargetFileView.WorkingExtensions =
					Dataset.Extensions;
			}
		}

		#endregion
	}
}
