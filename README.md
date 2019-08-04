# RoutedIP Operator #

This is a Kubernetes operator designed to operate the RoutedIP address feature on a platform. Using the RoutedIP CRD, it can create, delete, assign and unassign routed IPs across nodes in a cluster. The use case is meant for Ingress controllers implemented using a ClusterIP service with a selector for Pods. This is meant for the NodePort option of the service or ContainerPort.hostPort ports with a custom Cloud Firewall to allow the Ports on DigitalOcean class for Floating IPs.

## Planning ##

Current considerations:

RoutedIP controller secondary watch is Service to factor out Node Drain/Upgrade, Endpoint changes et al
ClusterRoutedIP class used to customize deployment 

* Operation considerations
    - On `ClusterRoutedIPClass` reconcile
        * On ClusterRoutedIPClass change
            - controllerutil.CreateOrUpdate with spec
            - ? Might need to catch in-process Job cases with DeepEqual changes
        * On JobSpec change // I don't think this needs to be watched, 
            - How to determine whether this is the updateRoutedIP or updateFirewall JobSpec?
            - controllerutil.CreateOrUpdate ClusterRoutedIPClass with JobSpec
        * On Job change
            - On updateRoutedIP success complete
                * clean up Job
                * ensure Firewall
                    - create updateFirewall Job
            - On failed complete
                * generate Event from Job failure
                * clean up Job
                * create new Job
    - On `RoutedIP` reconcile
        * On RoutedIP change
            - controllerutil.CreateOrUpdate with spec
        * On Service change
            - Check if match RoutedIPList serviceRef
            - ensure RoutedIP
                * get ClusterRoutedIPClass instance by className
                * create Job for updateRoutedIP owned by ClusterRoutedIPClass
    - Ensure RoutedIP ( with RoutedIP resource )
            * Build prospective Node list by RoutedIP.serviceRef
            * Check eligibility for Prospective Nodes
            * On multiple nodes eligible, elect youngest Node
            * Add taint on elected Node
            * If not `firewallPostAllow`, add Node to port on Firewall
            * Assign RoutedIP to elected Node
            * Remove taint from current Node
    - Ensure firewall
        * Build map of ServicePort.port/containerPort.hostPort with Nodes for ports outside NodePort range from RoutedIPs
        * Resolve nodes to Droplet IDs
        * Set ports with Droplet IDs on Firewall

* CRD `ClusterRoutedIPClass` Cluster scope
    - Can store `className` for RoutedIP implementation
    - Can store `firewallPostAllow` to patch firewalls only after assignment Default: false
    - Can store `nodePortRange` string for configured Node port range, since it's not discoverable via API Default: 30000-32767
    - Can store `perNodeIPLimit` for the configured Node <-> routed IP limit Default : 1
    - Can store `fromSecretsRef` Secret references for implementations values
    - Can store `updateRoutedIP` Job spec for RoutedIP implementation
    - Can store `updateFirewall` Job spec for Firewall implementation

* CRD `RoutedIP` Namespaced
    - Can store `className` RoutedIPClass that will handle the request
    - Can store `routedIP` assignable IP address
    - Can store `serviceRef` ( serviceName, namespace )
    - Can store status `assignedNode` to reference the Node that currently holds the IP
