# Testing strategy for module 2
#### Authors: Jianghong Wan, Zhishan Chen

To test the functionality of our tritonhttp server, we created this test section
implementing in python to send the input files to the server, and save the repsponses to the output files.

Table of Content:
#### 1. Basic functionality for 200 responses
 
dir_test.input - dir_test.output:   
To test if the server supports directories and subdirectories of the DocRoot, and each directory should be mapped to its index.html file. The server is expected to return correct header: Server, Last-Modified, Content-Type, Content-Length, Connection(if needed) and correct html file body.

mime_test.input - mime_test.output:   
Check if the response Content-Type Header is correct. In the test case, there are three requests with known file extensions in the MimeMap, and one request with extension .giaogiao, which is not recognized by the MimeMap. The Content-Type Headers of the output should be text/html, image/jpeg, image/png and application/octet-stream.

#### 2. Handle 404 responses
404_test.input - 404_test.output:   
This test case has three consecutive requests with bad paths, testing if a 404 error would be raised if the server is given a non-existent filename, a bad directory path, or a path that escapes the DocRoot. The server is expected to return three consecutive 404 errors.

#### 3. Handle 400 responses
400_test.input  - 400_test.output:    
The test contains an incomplete request which ends with "\r\n" instead of "\r\n\r\n". The server is expected to send back a 400 response after waiting for 5 seconds, and close the connection.

malformed_test1.input - malformed_test1.output:    
To test a malformed URL that doesn't start with "/".

malformed_test2.input - malformed_test2.output:    
To test a wrong method instead of "GET".

malformed_test3.input - malformed_test3.output:    
To test a wrong protocol instead of "HTTP/1.1".

malformed_test4.input - malformed_test4.output:    
To test a malformed initial line missing the protocol.

malformed_test5.input - malformed_test5.output:   
To test a malformed request missing host.

malformed_test6.input - malformed_test6.output:    
To test a malformed key-value pair.

#### 4. Concurrency tests
concurrency1.output - concurrency2.output:    
Outputs of concurrently connecting two clients to the server and sending requests at the same time. The inputs are written to two different files(test.py and test_concurrency.py), and the server is expected to handle the connection concurrently.

#### 5. Pipelining tests
pipeline_test1.input - pipeline_test1.output:    
This test case contains 5 pipelined valid requests, including 404 errors. The last request closes the connection by setting the "Connection:close" header.

pipeline_test2.input - pipeline_test2.output:     
This test case contains 4 pipelined valid requests and the 3rd request sets the "Connection:close" header. The server is expected to close the connection after executing the 3rd request and not handling the 4th request.

pipeline_test3.input - pipeline_test3.output:   
This test case contains 4 pipelined requests and the 3rd one is malformed. The server is expected to close the connection and send back a 400 response for the 3rd request.

twoTimed.input - twoTimed.output:   
timeout_test1.input - timeout_test1.output:   
Using test_timeout.py to send consecutive valid requests that do not properly set the “Connection: close” header to the server, and set a timer that counts how many seconds would the connection close after no more data is to send to the server.




