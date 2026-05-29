# Abstract
With the rapid acceleration of information through LLM evolution, keeping pace with pioneering technologies is increasingly difficult. While RSS (Really Simple Syndication) was developed to aggregate updates, synthesizing data from disparate sources remains a daunting task. This project explores the feasibility of leveraging local small language models (SLMs) to perform automated data crawling, synthesis, and illustration.

# Introduction 

# Background

# Goal
The flow chart of the artifact is illustrated as follow.
![architecture](architecture.png)
# Contribution

# Related Work

# Methodology
Monolithic System architecture.
## Subscriber
## Parser

# Expected problems and solution
## Format Consistency
Publisher content often varies in topic and structure. This fuzziness will be addressed via two methods:
1. Regex for keyword extraction.
2. NLP for classification.

## Content Consistency
Not all publishers strictly adhere to RSS protocols. This necessitates the development of modular formatters tailored to specific platforms or event publishers.

## LLM 
Unlike specialized services (e.g., NotebookLLM), many local models struggle with complex presentation. This system will feature a "hot-swappable" architecture, allowing users to toggle between local models or cloud APIs, though the primary focus remains maintaining a local-first environment.

# Expected results
A complete system with ability to actively gather information and news from the publisher, and perform data synthesis, and present data to the user in various form.