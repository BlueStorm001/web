# 最快的 web server
	//wrk -c400 -t12 -d30 --latency http://127.0.0.1:8086/search
	//******************4 vCPU 8 GiB*CPU 190%********************************
	//Running 30s test @ http://127.0.0.1:9527
	//12 threads and 400 connections
	//Thread Stats   Avg      Stdev     Max   +/- Stdev
	//	Latency     2.52ms    3.33ms  45.10ms   88.55%
	//	Req/Sec    19.43k     4.61k   44.49k    69.33%
	//	Latency Distribution
	//50%    1.20ms
	//75%    3.05ms
	//90%    6.39ms
	//99%   16.25ms
	//6965045 requests in 30.08s, 358.69MB read
	//Requests/sec: 231572.99
	//Transfer/sec:   11.93MB
	//*********************************************************************
	s := Service()
	s.Request = func(request proto3.Request) *proto3.Request {
		request.Header = nil
		request.Body = []byte("hello")
		return &request
	}
	s.NoHeader = true
	s.SyncResponse = true
	s.StartRun("9527")
