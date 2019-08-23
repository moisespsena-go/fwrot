// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import "strings"

var longHelp = strings.ReplaceAll(`Starts file writer rotation reads IN and writes to OUT.

EXAMPLES:
	NOTE: duration as minutely

	A. Basic example
		$ my_program | glogrotate -d m -o program.log
		$ my_program 2>&1 | glogrotate -d m -o program.log
	
	B. Input is STDIN, UDP, TCP and HTTP server
		main terminal:
			$ echo message from stdin | 
				glogrotate -d m -o program.log -i +udp:localhost:5678+tcp:localhost:5679

		secondary terminal:
			a. send message from UDP client:
				$ echo "message from UDP client" >/dev/udp/localhost/5678

			b. send message from TCP client:
				$ echo "message from TCP client " >/dev/udp/localhost/5679

			c. send message from HTTP client:
				$ curl -X POST -d "message from HTTP client" http://localhost:5680

	C. Input is STDIN and UDP server
		main terminal:
			$ (while true; do date; sleep 3; done) | 
				glogrotate -d m -o program.log -i +udp:localhost:5678

		secondary terminal - send message from UDP client:
			$ echo "date from UDP client: "$(date) >/dev/udp/localhost/5678

IN:
	Accept multiple inputs of STDIN, UDP and TCP servers.
	NOTE: Use plus char to join multiple values.
	      The first plus char, combines with STDIN.

	SERVERS:
		UDP: udp:ADDR, udp4:ADDR, udp6:ADDR ('udp:' is alias of 'udp4:')
			Max message size is 1024 bytes.

			Example:
				udp:localhost:5678
				udp4:localhost:5678
				udp:[::1]:5678
				udp6:[::1]:5678

		TCP: tcp:ADDR ('tcp:' is alias of 'tcp4:')
			Example:
				tcp:localhost:5679
				tcp4:localhost:5679
				tcp:[::1]:5679
				tcp6:[::1]:5679

		HTTP: http:ADDR ('http:' is alias of 'http4:')
			- Accept HTTP POST method and copy all request body.
			- Accept WebSocket INPUT on "/" and copy all message body.

			Example:
				http:localhost:5680
				http4:localhost:5680
				http:[::1]:5680
				http6:[::1]:5680
	
	Examples:
		1. Multiple servers
			udp:localhost:5678+tcp:localhost:5679+http:localhost:5680
		2. Multiple servers with STDIN
			+udp:localhost:5678+tcp:localhost:5679+http:localhost:5680

ENV VARIABLES:
	{N}_OUT, {N}_IN
	{N}_HISTORY_DIR, {N}_HISTORY_PATH, {N}_HISTORY_COUNT 
    {N}_DURATION, {N}_MAX_SIZE  
    {N}_DIR_MODE, {N}_FILE_MODE
	{N}_SILENT

	SET ENV variables to set default flag values.

	Usage example:
		Set duration as minutely and enable silent mode:
		$ export {N}_DURATION=m
		$ export {N}_SILENT=true
		
		run first program as background:
		$ my_first_program | glogrotate -d m -o first_program.log &

		run second program:
		$ my_second_program | glogrotate -d m -o second_program.log		
	
TIME FORMAT:
	%Y - Year. (example: 2006)
	%M - Month with left zero pad. (examples: 01, 12)
	%D - Day with left zero pad. (examples: 01, 31)
	%h - Hour with left zero pad. (examples: 00, 05, 23)
	%m - Minute with left zero pad. (examples: 00, 05, 59)
	%s - Second with left zero pad. (examples: 00, 05, 59)
	%Z - Time Zone. If not set, uses UTC time. (examples: +0700, -0330)
`, "{N}", strings.Trim(strings.ToUpper(name), "_"))
