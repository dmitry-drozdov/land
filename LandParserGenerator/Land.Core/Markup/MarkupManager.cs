﻿using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.IO;
using Newtonsoft.Json;
using Land.Core;
using Land.Core.Parsing.Tree;
using Land.Markup.Binding;
using Land.Markup.Relations;
using Land.Markup.Tree;
using Land.Markup.CoreExtension;

namespace Land.Markup
{
	public class MarkupManager
	{
		public MarkupManager(IContextFinder contextFinder)
		{
			ContextFinder = contextFinder;
			OnMarkupChanged += InvalidateRelations;
		}

		private RelationsManager Relations { get; set; } = new RelationsManager();

		public List<RelationNotification> TryGetRelations(out RelationsManager relationsManager)
		{
			if (IsValid)
			{
				if (!Relations.IsValid)
					Relations.RefreshCache(Markup);

				relationsManager = Relations;
				return Relations.CheckConsistency();
			}

			relationsManager = null;
			return new List<RelationNotification>();
		}

		public IContextFinder ContextFinder { get; set; }

		/// <summary>
		/// Коллекция точек привязки
		/// </summary>
		public ObservableCollection<MarkupElement> Markup = new ObservableCollection<MarkupElement>();

		/// <summary>
		/// Событие изменения разметки
		/// </summary>
		public event Action OnMarkupChanged;

		/// <summary>
		/// Очистка разметки
		/// </summary>
		public void Clear()
		{
			Markup.Clear();

			OnMarkupChanged?.Invoke();
		}

		/// <summary>
		/// Проверка того, что вся разметка синхронизирована с кодом
		/// </summary>
		/// <returns></returns>
		public bool IsValid => !GetLinearSequenceVisitor.GetPoints(Markup).Any(p => p.HasInvalidLocation);

		/// <summary>
		/// Помечаем отношения как нерелевантные относительно разметки
		/// </summary>
		public void InvalidateRelations()
		{
			Relations.IsValid = false;
		}

		/// <summary>
		/// Сброс узлов дерева у всех точек, связанных с указанным файлом
		/// </summary>
		public void InvalidatePoints(string fileName)
		{
			var stubNode = new Node("");
			stubNode.SetLocation(new PointLocation(0, 0, 0), new PointLocation(0, 0, 0));

			DoWithMarkup((MarkupElement elem) =>
			{
				if (elem is ConcernPoint concernPoint
					&& concernPoint.Context.FileName == fileName)
				{
					concernPoint.AstNode = stubNode;
					concernPoint.HasIrrelevantLocation = true;
				}
			});
		}

		/// <summary>
		/// Удаление элемента разметки
		/// </summary>
		public void RemoveElement(MarkupElement elem)
		{
			if (elem.Parent != null)
				elem.Parent.Elements.Remove(elem);
			else
				Markup.Remove(elem);

			OnMarkupChanged?.Invoke();
		}

		/// <summary>
		/// Добавление функциональности
		/// </summary>
		public Concern AddConcern(string name, string comment = null, Concern parent = null)
		{
			var concern = new Concern(name, comment, parent);
			AddElement(concern);

			OnMarkupChanged?.Invoke();
			return concern;
		}

		/// <summary>
		/// Добавление точки привязки
		/// </summary>
		public ConcernPoint AddConcernPoint(TargetFileInfo sourceInfo, string name = null, string comment = null, Concern parent = null)
		{
			var point = new ConcernPoint(sourceInfo, parent);

			if (!String.IsNullOrEmpty(name))
				point.Name = name;
			point.Comment = comment;

			AddElement(point);

			OnMarkupChanged?.Invoke();
			return point;
		}

		/// <summary>
		/// Добавление всей "суши", присутствующей в дереве разбора
		/// </summary>
		public void AddLand(TargetFileInfo sourceInfo)
		{
			var visitor = new LandExplorerVisitor();
			/// При добавлении всей суши к разметке, в качестве целевого узла передаётся корень дерева
			sourceInfo.TargetNode.Accept(visitor);

			/// Группируем land-сущности по типу (символу)
			foreach (var group in visitor.Land.GroupBy(l => l.Symbol))
			{
				var concern = AddConcern(group.Key);

				/// В пределах символа группируем по псевдониму
				var subgroups = group.GroupBy(g => g.Alias);

				/// Для всех точек, для которых указан псевдоним
				foreach (var subgroup in subgroups.Where(s => !String.IsNullOrEmpty(s.Key)))
				{
					/// создаём подфункциональность
					var subconcern = AddConcern(subgroup.Key, null, concern);

					foreach (var point in subgroup)
					{
						sourceInfo.TargetNode = point;
						AddElement(new ConcernPoint(sourceInfo, subconcern));
					}
				}

				/// Остальные добавляются напрямую к функциональности, соответствующей символу
				var points = subgroups.Where(s => String.IsNullOrEmpty(s.Key))
					.SelectMany(s => s).ToList();

				foreach (var point in points)
				{
					sourceInfo.TargetNode = point;
					AddElement(new ConcernPoint(sourceInfo, concern));
				}
			}

			OnMarkupChanged?.Invoke();
		}

		/// <summary>
		/// Получение всех узлов, к которым можно привязаться,
		/// если команда привязки была вызвана в позиции offset
		/// </summary>
		public LinkedList<Node> GetConcernPointCandidates(Node root, SegmentLocation selection)
		{
			var pointCandidates = new LinkedList<Node>();
			var currentNode = root;

			/// В качестве кандидатов на роль помечаемого участка рассматриваем узлы от корня,
			/// содержащие текущую позицию каретки
			while (currentNode != null)
			{
				if (currentNode.Options.IsSet(MarkupOption.LAND))
					pointCandidates.AddFirst(currentNode);

				currentNode = currentNode.Children
					.Where(c => c.Location != null && c.Location.Includes(selection))
					.FirstOrDefault();
			}

			return pointCandidates;
		}

		/// <summary>
		/// Смена узла, к которому привязана точка
		/// </summary>
		public void RelinkConcernPoint(ConcernPoint point, TargetFileInfo targetInfo)
		{
			point.Relink(targetInfo);

			OnMarkupChanged?.Invoke();
		}

		/// <summary>
		/// Смена узла, к которому привязана точка
		/// </summary>
		public void RelinkConcernPoint(ConcernPoint point, RemapCandidateInfo candidate)
		{
			point.Relink(candidate);

			OnMarkupChanged?.Invoke();
		}

		/// <summary>
		/// Получение списка точек привязки для текущего дерева разметки
		/// </summary>
		public List<ConcernPoint> GetConcernPoints()
		{
			return GetLinearSequenceVisitor.GetPoints(Markup);
		}

		/// <summary>
		/// Перемещение элемента разметки к новому родителю
		/// </summary>
		public void MoveTo(Concern newParent, MarkupElement elem)
		{
			if (elem.Parent != null)
				elem.Parent.Elements.Remove(elem);
			else
				Markup.Remove(elem);

			elem.Parent = newParent;

			if(newParent != null)
				newParent.Elements.Add(elem);
			else
				Markup.Add(elem);

			OnMarkupChanged?.Invoke();
		}

		public void Serialize(string fileName, bool useRelativePaths)
		{
			if (useRelativePaths)
			{
				/// Превращаем указанные в точках привязки абсолютные пути в пути относительно файла разметки
				var directoryUri = new Uri(Path.GetDirectoryName(fileName) + "/");
				DoWithMarkup((MarkupElement elem) =>
				{
					if (elem is ConcernPoint p)
					{
						p.Context.FileName = Uri.UnescapeDataString(
							directoryUri.MakeRelativeUri(new Uri(p.Context.FileName)).ToString()
						);
					}
				});
			}

			using (StreamWriter fs = new StreamWriter(fileName, false))
			{
				var unit = new SerializationUnit()
				{
					Markup = Markup,
					ExternalRelatons = Relations.ExternalRelations.GetRelatedPairs()
				};

				fs.Write(JsonConvert.SerializeObject(unit, Formatting.Indented));
			}

			if (useRelativePaths)
			{
				/// Трансформируем пути обратно в абсолютные
				DoWithMarkup((MarkupElement elem) =>
				{
					if (elem is ConcernPoint p)
					{
						p.Context.FileName = Path.GetFullPath(
							Path.Combine(Path.GetDirectoryName(fileName), p.Context.FileName)
						);
					}
				});
			}
		}

		public void Deserialize(string fileName)
		{
			Clear();

			using (StreamReader fs = new StreamReader(fileName))
			{
				var unit = JsonConvert.DeserializeObject<SerializationUnit>(fs.ReadToEnd(),
					new JsonSerializerSettings()
					{
						Converters = { new MarkupElementConverter() }
					});

				/// Фиксируем разметку
				Markup = unit.Markup;

				/// Восстанавливаем обратные связи между потомками и предками
				DoWithMarkup(e =>
				{
					if (e is Concern c)
					{
						foreach (var elem in c.Elements)
							elem.Parent = c;
					}
				});

				/// Запоминаем external-отношения между функциональностями
				Relations.RefreshElements(Markup);

				foreach (var pair in unit.ExternalRelatons)
					Relations.AddExternalRelation(pair.RelationType, pair.Item0, pair.Item1);
			}

			DoWithMarkup((MarkupElement elem) =>
			{
				if (elem is ConcernPoint p && !Path.IsPathRooted(p.Context.FileName))
				{
					p.Context.FileName = Path.GetFullPath(
						Path.Combine(Path.GetDirectoryName(fileName), p.Context.FileName)
					);
				}
			});
		}

		/// <summary>
		/// Поиск узла дерева, которому соответствует заданная точка привязки
		/// </summary>
		public List<RemapCandidateInfo> Find(ConcernPoint point, TargetFileInfo targetInfo)
		{
			return ContextFinder.Find(point, targetInfo);
		}

		/// <summary>
		/// Получение списка файлов, в которых есть точки привязки
		/// </summary>
		public HashSet<string> GetReferencedFiles()
		{
			return new HashSet<string>(
				GroupPointsByFileVisitor.GetGroups(Markup).Select(p => p.Key)
			);
		}

		#region Перепривязка

		/// <summary>
		/// Размер топа кандидатов, ранжированных по похожести, возвращаемого при неоднозначности
		/// </summary>
		public int AmbiguityTopCount { get; set; } = 10;

		/// <summary>
		/// Похожесть, ниже которой не рассматриваем элемент как кандидата
		/// </summary>
		public double GarbageThreshold { get; set; } = 0.4;

		public Dictionary<ConcernPoint, List<RemapCandidateInfo>> Remap(List<TargetFileInfo> targetFiles, bool useLocalRemap, bool allowAutoDecisions)
		{
			var ambiguous = useLocalRemap
				? LocalRemap(targetFiles, allowAutoDecisions)
				: GlobalRemap(targetFiles, allowAutoDecisions);

			OnMarkupChanged?.Invoke();

			return ambiguous;
		}

		private Dictionary<ConcernPoint, List<RemapCandidateInfo>> LocalRemap(List<TargetFileInfo> targetFiles, bool allowAutoDecisions)
		{
			var groupedByFile = GroupPointsByFileVisitor.GetGroups(Markup);
			var ambiguous = new Dictionary<ConcernPoint, List<RemapCandidateInfo>>();

			foreach(var fileGroup in groupedByFile)
			{
				var file = targetFiles.Where(f => f.FileName == fileGroup.Key).FirstOrDefault();

				if(file != null)
				{
					var groupedByType = fileGroup.Value.GroupBy(p => p.Context.NodeType).ToDictionary(g => g.Key, g => g.ToList());
					var groupedFiles = GroupNodesByTypeVisitor.GetGroups(file.TargetNode, groupedByType.Keys);

					var result = ContextFinder.Find(groupedByType, groupedFiles, file);

					foreach (var kvp in result)
					{
						var candidates = kvp.Value
							//.TakeWhile(c=>c.Similarity >= GarbageThreshold)
							.Take(AmbiguityTopCount).ToList();

						if (!allowAutoDecisions || !ApplyCandidate(kvp.Key, candidates))
							ambiguous[kvp.Key] = candidates;
					}
				}
				else
				{
					foreach (var point in fileGroup.Value)
						point.AstNode = null;
				}
			}

			return ambiguous;
		}

		private Dictionary<ConcernPoint, List<RemapCandidateInfo>> GlobalRemap(List<TargetFileInfo> targetFiles, bool allowAutoDecisions)
		{
			var ambiguous = new Dictionary<ConcernPoint, List<RemapCandidateInfo>>();

			/// Группируем точки привязки по типу помеченной сущности 
			var groupedPoints = GroupPointsByTypeVisitor.GetGroups(Markup);
			var accumulator = groupedPoints.SelectMany(e => e.Value).ToDictionary(e => e, e => new List<RemapCandidateInfo>());

			foreach (var file in targetFiles)
			{
				/// Группируем узлы AST файла, к которому попытаемся перепривязаться,
				/// по типам точек, к которым требуется перепривязка
				var groupedFiles = GroupNodesByTypeVisitor.GetGroups(file.TargetNode, groupedPoints.Keys);

				/// Похожести, посчитанные для сущностей из текущего файла
				var currentRes = ContextFinder.Find(groupedPoints, groupedFiles, file);

				foreach (var kvp in currentRes)
					accumulator[kvp.Key].AddRange(kvp.Value);
			}

			foreach (var kvp in accumulator)
			{
				var candidates = kvp.Value
					//.TakeWhile(c => c.Similarity >= GarbageThreshold)
					.Take(AmbiguityTopCount).ToList();

				if (!allowAutoDecisions || !ApplyCandidate(kvp.Key, candidates))
					ambiguous[kvp.Key] = candidates;
			}

			return ambiguous;
		}

		/// <summary>
		/// Перепривязка точки
		/// </summary>
		public Dictionary<ConcernPoint, List<RemapCandidateInfo>> Remap(ConcernPoint point, TargetFileInfo targetInfo)
		{
			var ambiguous = new Dictionary<ConcernPoint, List<RemapCandidateInfo>>();
			var candidates = ContextFinder.Find(point, targetInfo)
				.TakeWhile(c => c.Similarity >= GarbageThreshold)
				.Take(AmbiguityTopCount).ToList();

			if (!ApplyCandidate(point, candidates))
				ambiguous[point] = candidates;

			OnMarkupChanged?.Invoke();

			return ambiguous;
		}

		private bool ApplyCandidate(ConcernPoint point, IEnumerable<RemapCandidateInfo> candidates)
		{
			var first = candidates.FirstOrDefault();

			if (first?.IsAuto ?? false)
			{
				point.Context = first.Context;
				point.AstNode = first.Node;
				return true;
			}
			else
			{
				point.AstNode = null;
				return false;
			}
		}

		#endregion

		/// <summary>
		/// Обобщённое добавление элемента разметки
		/// </summary>
		/// <param name="elem"></param>
		private void AddElement(MarkupElement elem)
		{
			if (elem.Parent == null)
				Markup.Add(elem);
			else
				elem.Parent.Elements.Add(elem);
		}

		/// <summary>
		/// Совершение заданного действия со всеми элементами разметки
		/// </summary>
		public void DoWithMarkup(Action<MarkupElement> action)
		{
			foreach (var elem in Markup)
				DoWithMarkupSubtree(action, elem);
		}

		/// <summary>
		/// Совершение заданного действия со всеми элементами поддерева разметки
		/// </summary>
		private void DoWithMarkupSubtree(Action<MarkupElement> action, MarkupElement root)
		{
			var elements = new Queue<MarkupElement>();
			elements.Enqueue(root);

			while (elements.Count > 0)
			{
				var elem = elements.Dequeue();

				if (elem is Concern concern)
					foreach (var child in concern.Elements)
						elements.Enqueue(child);

				action(elem);
			}
		}
	}
}