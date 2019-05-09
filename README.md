nexthop-checker
===============

Nexthop-checker loads all NextHops and Flows from vRouter Introspection service and compares ID of nexthop with existing nexthops.

If ID of nexthop from flow doesn't exist, It will be reported to console without extra parametes If NH exists there is no information reported
to console by default. For more verbose level use `-v`  flag which is described below. 

Important flags:

<pre>
--host      IP or Hostname of vRouter introspect (default: localhost)
--port      Port of introspect service (default: 8085)
--progress  Shows progress bar of checking process
--h         Print help, list of all flags
--v         Verbose level (1,2,3)
</pre>

 Verbose levels:
 * Error          --v 0
 * Warning        --v 1 (default)
 * Info           --v 2
 * Debug mode     --v 3 
