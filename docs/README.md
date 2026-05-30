# Abstract
With the rapid acceleration of information driven by the evolution of Large Language Models (LLMs), keeping pace with pioneering technologies has become increasingly difficult. While Really Simple Syndication (RSS) was developed to aggregate updates, synthesizing meaningful insights from disparate sources remains a daunting task. This project explores the feasibility of leveraging local Small Language Models (SLMs) to perform automated data crawling, synthesis, and structured technical illustration.

# Introduction 

# Background

# Goal
The primary objective of this project is to aggregate information from diverse publishers within a specific technical domain and automatically synthesize the findings into a structured technical briefing slide deck.

## Topic
Artificial Intelligence, specifically Machine Learning (ML), serves as the primary target domain for this project. The publishers in this space can be broadly categorized into two types: academia and industry. These two spheres exhibit distinct characteristics in their technical depth, focus, and contributions to state-of-the-art AI technology.

## Source/Publisher
The publisher for the topics can be vary, especially in industry type. 
- Academy
     - arxiv
- Industry:
    - Corporate research blogs
    - Individual technical publishers

While academic research traditionally falls under established publishers like IEEE, SpringerLink, ScienceDirect, and ACM, much of this literature sits behind paywalls. Thanks to the open-access culture of the AI research community, arXiv provides a robust, free repository of state-of-the-art research papers.

On the industry side, contributions primarily come from major corporate research labs (e.g., Google, NVIDIA, Meta). However, relying solely on corporate blogs risks missing exceptional insights from independent researchers. To bridge this gap, this project includes individual publishers across various platforms, including X (formerly Twitter), Threads, and technical blogging sites like Medium.

## Feature
The core feature of this system is its ability to synthesize cross-platform information into a cohesive, presentation-ready slide format (often in .pptx format). The system will support customizable reporting intervals, such as weekly summaries ("top developments in ML over the past 7 days"), monthly roundups, or real-time latest updates.

Why slides? During technical briefings, slides are the standard medium for conveying complex data. Breaking down technical insights into structured, visual layers makes information significantly easier to digest and follow compared to dense, plain text.

The goal of these generated slides is not to replace deep reading, but to act as a high-signal discovery tool. The system aims to synthesize just enough context to spark a reader's interest, allowing them to think: "I am aware of this development, I understand the problem they are solving, their proposed approach, and the potential impact. I don't know every implementation detail yet, but now I know exactly where to dig deeper if it applies to my current challenges." This design reflects a core philosophy: focusing the system on identifying actionable solutions rather than merely cataloging technical concepts.


# Contribution

# Related Work
[Feedify](https://feedify.net/)
# Methodology
Monolithic System architecture.

Golang + Wails (webui) + React Typescript

## Subscriber
## Parser

# Expected problems and solution
## Format Consistency
- Problem: Publisher content varies wildly in structure, formatting, and thematic focus.
- Solution: This structural fuzziness will be addressed using a hybrid approach:
  - Regular Expressions (Regex): For deterministic keyword and metadata extraction.
  - Natural Language Processing (NLP): For semantic classification and thematic grouping.

## Content Consistency
- Problem: Not all digital publishers strictly adhere to or maintain standard RSS protocols, leading to broken feeds or missing data.
- Solution: The system will utilize a modular architecture, featuring decoupled, platform-specific formatters tailored to scrape and normalize data from non-standard platforms.

## LLM 
- Problem: Unlike specialized cloud services (e.g., NotebookLM), local small language models (SLMs) often struggle with complex structured outputs and presentation formatting.
- Solution: The system will implement a "hot-swappable" inference architecture. This allows users to seamlessly toggle between local SLMs and cloud-based APIs depending on their compute availability, though the core development focus will remain on maintaining a private, local-first environment.

# Expected results
A complete, end-to-end system capable of actively gathering news from academic and industry publishers, synthesizing disparate data points, and presenting clear, structured insights to the user across various formats (including automated slide decks).
