# LocalReader

`LocalReader` is the file-based asset retrieval layer responsible for reading raw payloads, text summaries, and image slides directly from the local filesystem.

Unlike database Readers, it exclusively handles file I/O using defined directory patterns and numerical sequencing rules. It does not perform database operations or business logic.

The reader is responsible for:

* Reading raw RSS XML payloads from disk.
* Locating and loading text summaries for Topics and Posts.
* Scanning, sorting, and loading ordered PNG slide images from asset directories.
* Encapsulating directory structure pathing logic.

# RSS Operations

## ReadRss

Retrieves the raw RSS XML content for a specific RSS ID.

### Signature

```go
ReadRss(ID int64) (string, error)
```

### Returns

| Type     | Description                                   |
| -------- | --------------------------------------------- |
| `string` | The raw XML payload content.                  |
| `error`  | File read failure or non-existent path error. |

### Behavior

* Constructs path `{basePath}/rss/{ID}.xml`.
* Reads and returns the raw file contents as a string.

# Topic Operations

## ReadTopicSummary

Retrieves the generated text summary for a specific Topic using its ID and RSS hash.

### Signature

```go
ReadTopicSummary(ID int64, rssHash string) (string, error)
```

### Returns

| Type     | Description                                         |
| -------- | --------------------------------------------------- |
| `string` | Text content of the topic summary.                  |
| `error`  | File read failure or missing summary payload error. |

### Behavior

* Constructs directory name `{ID}_{rssHash}`.
* Reads file at `{basePath}/summary/topic/{ID}_{rssHash}/summary.txt`.

## ReadTopicSlide

Retrieves all PNG slide image files associated with a Topic, sorted numerically by filename.

### Signature

```go
ReadTopicSlide(ID int64, rssHash string) ([][]byte, error)
```

### Returns

| Type       | Description                                      |
| ---------- | ------------------------------------------------ |
| `[][]byte` | An ordered slice of image raw byte arrays.       |
| `error`    | Directory read error or file extraction failure. |

### Behavior

* Targets directory `{basePath}/summary/topic/{ID}_{rssHash}/slide`.
* Filters for `.png` files with numerical filenames, such as `1.png` and `2.png`.
* Ignores non-PNG files, directories, and non-numeric filenames.
* Returns image bytes sorted in ascending numerical order.

# Post Operations

## ReadPostSummary

Retrieves the historical target summary text for a given Post ID.

### Signature

```go
ReadPostSummary(ID int64) (string, error)
```

### Returns

| Type     | Description                                      |
| -------- | ------------------------------------------------ |
| `string` | Text content of the post summary.                |
| `error`  | File read failure or missing post summary error. |

### Behavior

* Reads file at `{basePath}/summary/post/{ID}/summary.txt`.

## ReadPostSlide

Retrieves all PNG slide image files associated with a Post ID, sorted numerically by filename.

### Signature

```go
ReadPostSlide(ID int64) ([][]byte, error)
```

### Returns

| Type       | Description                                      |
| ---------- | ------------------------------------------------ |
| `[][]byte` | An ordered slice of image raw byte arrays.       |
| `error`    | Directory read error or file extraction failure. |

### Behavior

* Targets directory `{basePath}/summary/post/{ID}/slide`.
* Filters for `.png` files with numerical filenames.
* Ignores non-PNG files, directories, and non-numeric filenames.
* Returns image bytes sorted in ascending numerical order.
