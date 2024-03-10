using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.IO;
using System.Runtime.Serialization;
using System.Diagnostics;
using System.Runtime.CompilerServices;

namespace Land.Core.Parsing
{
	[DataContract]
	public class Statistics
	{
		[DataMember]
		public int CharsCount { get; set; }
		[DataMember]
		public int TokensCount { get; set; }
		[DataMember]
		public TimeSpan GeneralTimeSpent { get; set; }
		[DataMember]
		public TimeSpan RecoveryTimeSpent { get; set; }
		[DataMember]
		public int RecoveryTimes { get; set; }
		[DataMember]
		public int RecoveryTimesAny { get; set; }
		[DataMember]
		public int LongestRollback { get; set; }

		public static Statistics operator +(Statistics a, Statistics b)
		{
			return new Statistics
			{
				TokensCount = a.TokensCount + b.TokensCount,
				GeneralTimeSpent = a.GeneralTimeSpent + b.GeneralTimeSpent,
				RecoveryTimeSpent = a.RecoveryTimeSpent + b.RecoveryTimeSpent,
				LongestRollback = a.LongestRollback + b.LongestRollback,
				RecoveryTimes = a.RecoveryTimes + b.RecoveryTimes,
				RecoveryTimesAny = Math.Max(a.RecoveryTimesAny, b.RecoveryTimesAny)
			};
		}

		public override string ToString()
		{
			return $"Количество токенов: {TokensCount};{Environment.NewLine}" +
				$"Время парсинга: {GeneralTimeSpent.ToString(@"hh\:mm\:ss\:ff")};{Environment.NewLine}" +
				$"Время восстановлений от ошибки: {RecoveryTimeSpent.ToString(@"hh\:mm\:ss\:ff")};{Environment.NewLine}" +
				$"Количество восстановлений от ошибки: {RecoveryTimes};{Environment.NewLine}" +
				$"Восстановлений при сопоставлении Any: {RecoveryTimesAny};{Environment.NewLine}" +
				$"Количество токенов в самом длинном возврате: {LongestRollback}{Environment.NewLine}";
		}
	}

	public class Durations
	{
		public Dictionary<string, long> Stats { get; private set; } = new Dictionary<string, long>();

		[MethodImpl(MethodImplOptions.AggressiveInlining)]
		public void Start()
		{
			watch = Stopwatch.StartNew();
		}

		[MethodImpl(MethodImplOptions.AggressiveInlining)]
		public void Stop(string method)
		{
			watch.Stop();
			if (Stats.ContainsKey(method))
				Stats[method] += watch.ElapsedMilliseconds;
			else
				Stats.Add(method, watch.ElapsedMilliseconds);
		}

		[MethodImpl(MethodImplOptions.AggressiveInlining)]
		public void Add(string method, Stopwatch watch)
		{
			watch.Stop();
			if (Stats.ContainsKey(method))
				Stats[method] += watch.ElapsedMilliseconds;
			else
				Stats.Add(method, watch.ElapsedMilliseconds);
		}

		public static Durations operator +(Durations a, Durations b)
		{
			// add new values (b)
			foreach (var key in b.Stats.Keys)
			{
				if (a.Stats.ContainsKey(key))
					a.Stats[key] += b.Stats[key];
				else
					a.Stats.Add(key, b.Stats[key]);
			}

			return a;
		}

		private Stopwatch watch = new Stopwatch();
	}
}
