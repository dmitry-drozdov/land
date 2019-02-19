﻿using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Media;

using Microsoft.Win32;

using Land.Core;
using Land.Core.Parsing;
using Land.Core.Parsing.Tree;
using Land.Core.Parsing.Preprocessing;
using Land.Core.Markup;
using Land.Control.Helpers;

namespace Land.Control
{
	public partial class LandExplorerControl : UserControl
	{
		private void RelationParticipants_Swap_Click(object sender, RoutedEventArgs e)
		{
			var tmp = RelationSource.Tag;

			RelationSource.Tag = RelationTarget.Tag;
			RelationTarget.Tag = tmp;

			RefreshRelationCandidates();
		}

		private void RelationSource_Reset_Click(object sender, RoutedEventArgs e)
		{
			RelationSource.Tag = null;
			RelationCandidatesList.ItemsSource = null;
		}

		private void RelationTarget_Reset_Click(object sender, RoutedEventArgs e)
		{
			RelationTarget.Tag = null;
			RelationCandidatesList.ItemsSource = null;
		}

		private void RelationCancelButton_Click(object sender, RoutedEventArgs e)
		{
			RelationSource_Reset_Click(null, null);
			RelationTarget_Reset_Click(null, null);

			SetStatus("Добавление отношения отменено", ControlStatus.Ready);
		}

		private void RelationSaveButton_Click(object sender, RoutedEventArgs e)
		{
			SetStatus("Отношение добавлено", ControlStatus.Success);
		}

		private void RefreshRelationCandidates()
		{

		}
	}
}