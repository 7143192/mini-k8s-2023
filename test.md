# Test Info
## How to do test ?(lyh_iter2 version)
1. start one terminal, run "make".
2. start "etcd" and "cadvisor (at port 8090)" in your test machine.
2. run "make master_start" in the same terminal (or a new one. )
3. run "make node_start" in a new terminal.
4. start a new terminal to run "make test", this should pass all tests for now.
 