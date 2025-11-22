using Jaeger.Reporters;
using Jaeger.Samplers;
using Jaeger.Senders.Thrift;
using Jaeger;
using OpenTracing.Util;
using OpenTracing;

namespace Land.Core.Core.Parsing
{
	public static class Tracing
	{
		public static void Init()
		{
			if (GlobalTracer.IsRegistered())
				return;

			var sender = new UdpSender("localhost", 6831, 0);
			var reporter = new RemoteReporter.Builder()
			    .WithSender(sender)
			    .Build();

			var tracer = new Tracer.Builder("Praser")
			    .WithReporter(reporter)
			    .WithSampler(new ConstSampler(true))
			    .Build();


			GlobalTracer.Register(tracer);
		}

		public static ITracer Tracer => GlobalTracer.Instance;
	}
}
