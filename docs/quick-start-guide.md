# Quick Start Guide

## Intended Audience

This documentation is written for users who are already familiar with one or more of the ASCE core applications such as Hub or Data Tool and who are starting to work with data bottles.

## Overview

The ASCE Telemetry Server automates longitudinal tracking of data sets, known as data bottles, that are created using ASCE Data Tool and that may be launched and used in ASCE Hub.

Users of Data Tool can (optionally) configure their installation to automate the tracking of data bottle usage and registry locations via the ASCE Telemetry Server. Leveraging this integration is a way of conducting responsible AI (rAI) research and supports best practices in data science. Additionally, integrating the ASCE Telemetry Server into a Data Tool installation makes data bottles of potential interest to other researchers discoverable. In turn, researchers can access the ASCE Telemetry Server's graphical user interface (GUI) to search for data sets by using selectors to query the metadata associated with bottles.

Thus, the ASCE Telemetry Server helps solve problems of *data discoverability* and *data provenance* for ACT3 researchers.

### Use Cases

When data bottles contain machine learning models tuned for a certain task, it is especially useful and important to address problems associated *data discoverability* and *data provenance*.

When Data Tool users add labels to document metadata associated with a bottle, metrics such as "accuracy" make it possible for a researcher or research team to use the Telemetry Server's leaderboard to easily identify which model currently performs best for a given task, inspect it, and try to improve upon the work.

### Integrations

The ASCE Telemetry server is designed to be integrated with other ASCE core applications including:

- Data Tool
- Hub

These integrations are supported and facilitated by the ACT3 Login script.
