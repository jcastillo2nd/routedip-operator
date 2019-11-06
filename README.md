# RoutedIP Operator #

This is a Kubernetes operator designed to operate the RoutedIP address feature on a platform. Using the RoutedIP CRD, it can create, delete, assign and unassign routed IPs across nodes in a cluster. The use case is meant for Ingress controllers implemented using a ClusterIP service with a selector for Pods in concert with a Platform that supports routable IP addresses like the DigitalOcean Floating IP feature.

## RoutedIP CRD ##

The RoutedIP resource is the IP address that will be assigned across nodes. This can be created through a RoutedIPIssuer, or can be recovered by defining the address in the RoutedIP spec.

## RoutedIPIssuer CRD ##

The RoutedIPIssuer resource defines the API functionality for managing a RoutedIP. A RoutedIPIssuer is required in order for a RoutedIP to operate.

## Implementation Details ##

A RoutedIPIssuer client supports

- Finding a RoutedIP address ( FindRoutedIP )

- Creating a RoutedIP address ( CreateRoutedIP )

- Deleting a RoutedIP address ( DeleteRoutedIP <- routedIPReclaimPolicy {*'Retain'*, 'Delete'})

- Electing an eligible Node for RoutedIP assignment ( ElectRoutedIPNode )

- Assigning a RoutedIP to a Node ( AssignRoutedIP )

- Updating a firewall for RoutedIP operation ( RoutedIPFirewallUpdate <- routedIPFirewallName )

  - Finding a Firewall configuration ( FindRoutedIPFirewall )

  - Finding ports used in RoutedIP service ( FindRoutedIPPorts )

  - Updating Firewall port configuration ( UpdateRoutedIPFirewallPorts )

    - Adding a Firewall Port Rule ( AddRoutedIPFirewallPort )

    - Removing a Firewall Port Rule ( DeleteRoutedIPFirewallPort )

  - Updating a Firewall node configuration ( UpdateRoutedIPFirewallNode )

    - Adding a Node Firewall rule ( AddRoutedIPFirewallNode )

    - Removing a Node from a firewall ( DeleteRoutedIPFirewallNode )

The order for the update should occur in a manner to maximize service availability

ElectRoutedIPNode - Pick a node that will now hold the RoutedIP
FindRoutedIPFirewall - Ensure we have a Firewall to work on, if at all
FindRoutedIPPorts - Ensure we have a list of the required ports
AddRoutedIPFirewallPort - Ensure Port is added to Firewall
AddRoutedIPFirewallNode - Ensure Node is added to Firewall
AssignRoutedIP - Ensure RoutedIP is moving traffic to elected Node
DeleteRoutedIPFirewallPort - Ensure extraneous Ports are removed from Firewall
DeleteRoutedIPFirewallNode - Ensure extraneous Nodes are removed from Firewall


