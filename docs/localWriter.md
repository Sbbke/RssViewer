# LocalWriter

`LocalWriter` is the file-based asset persistence layer responsible for writing, updating, and deleting raw payloads, text summaries, and slide images directly on the local filesystem.

Unlike database Writers, it exclusively handles file I/O using defined directory structures and numerical sequencing rules. It does not perform database operations or business logic.

The writer is responsible for:

* Creating and storing raw RSS XML payloads on disk.
* Creating and storing text summaries for Topics and Posts.
* Creating and storing ordered PNG slide images for Topics and Posts.
* Updating existing RSS payloads and text summaries using atomic file replacement.
* Updating slide assets using atomic directory replacement.
* Deleting RSS, summary, and slide artifacts from local storage.
* Encapsulating filesystem path construction and storage management logic.

## Initialization

### NewLocalWriter

Creates a new `LocalWriter` using the specified base storage directory.

### Signature

```go
NewLocalWriter(folder_path string) *LocalWriter
```

### Parameters

| Name          | Type     | Description                                           |
| ------------- | -------- | ----------------------------------------------------- |
| `folder_path` | `string` | Base directory used for all local storage operations. |

### Returns

| Type           | Description                                                          |
| -------------- | -------------------------------------------------------------------- |
| `*LocalWriter` | A new `LocalWriter` configured with the specified base storage path. |

# RSS Operations

## CreateRss

Creates or overwrites the raw RSS XML payload for a specific RSS ID.

### Signature

```go
CreateRss(ID int64, content string) error
```

### Parameters

| Name      | Type     | Description                   |
| --------- | -------- | ----------------------------- |
| `ID`      | `int64`  | Unique RSS identifier.        |
| `content` | `string` | Raw RSS XML content to store. |

### Returns

| Type    | Description                               |
| ------- | ----------------------------------------- |
| `error` | Directory creation or file write failure. |

### Behavior

* Constructs directory `{basePath}/rss/{ID}`.
* Creates the directory if it does not already exist.
* Writes the RSS content to `{basePath}/rss/{ID}/rss.xml`.
* Creates or overwrites the target file with file permissions `0644`.

## UpdateRss

Atomically updates the raw RSS XML payload for a specific RSS ID.

### Signature

```go
UpdateRss(ID int64, content string) error
```

### Parameters

| Name      | Type     | Description              |
| --------- | -------- | ------------------------ |
| `ID`      | `int64`  | Unique RSS identifier.   |
| `content` | `string` | New raw RSS XML content. |

### Returns

| Type    | Description                                                                           |
| ------- | ------------------------------------------------------------------------------------- |
| `error` | Parent directory creation, temporary file, synchronization, close, or rename failure. |

### Behavior

* Targets `{basePath}/rss/{ID}/rss.xml`.
* Ensures the parent directory exists.
* Writes content to a temporary file in the same directory.
* Synchronizes the temporary file to disk using `Sync()`.
* Closes the temporary file.
* Renames the temporary file to the target path.
* Removes the temporary file if the update fails before the rename operation completes.

This prevents a partially written RSS file from replacing the existing payload.

## DeleteRss

Deletes all local RSS storage associated with a specific RSS ID.

### Signature

```go
DeleteRss(ID int64) error
```

### Parameters

| Name | Type    | Description            |
| ---- | ------- | ---------------------- |
| `ID` | `int64` | Unique RSS identifier. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Targets `{basePath}/rss/{ID}`.
* Removes the entire RSS directory and its contents.
* Returns `nil` if the target does not already exist.

# Post Operations

## CreatePostSummary

Creates or overwrites the summary text for a specific Post ID.

### Signature

```go
CreatePostSummary(ID int64, content string) error
```

### Parameters

| Name      | Type     | Description             |
| --------- | -------- | ----------------------- |
| `ID`      | `int64`  | Unique Post identifier. |
| `content` | `string` | Summary text to store.  |

### Returns

| Type    | Description                               |
| ------- | ----------------------------------------- |
| `error` | Directory creation or file write failure. |

### Behavior

* Constructs directory `{basePath}/summary/post/{ID}`.
* Creates the directory if it does not already exist.
* Writes the summary to `{basePath}/summary/post/{ID}/summary.txt`.
* Creates or overwrites the target file with file permissions `0644`.

## UpdatePostSummary

Atomically updates the summary text for a specific Post ID.

### Signature

```go
UpdatePostSummary(ID int64, content string) error
```

### Parameters

| Name      | Type     | Description             |
| --------- | -------- | ----------------------- |
| `ID`      | `int64`  | Unique Post identifier. |
| `content` | `string` | New summary text.       |

### Returns

| Type    | Description                                                                           |
| ------- | ------------------------------------------------------------------------------------- |
| `error` | Parent directory creation, temporary file, synchronization, close, or rename failure. |

### Behavior

* Targets `{basePath}/summary/post/{ID}/summary.txt`.
* Uses atomic temporary-file replacement.
* Synchronizes the temporary file before replacing the target.
* Removes the temporary file if the operation fails.

## CreatePostSlide

Creates or overwrites the PNG slide assets associated with a Post ID.

### Signature

```go
CreatePostSlide(ID int64, images [][]byte) error
```

### Parameters

| Name     | Type       | Description                                  |
| -------- | ---------- | -------------------------------------------- |
| `ID`     | `int64`    | Unique Post identifier.                      |
| `images` | `[][]byte` | Ordered collection of raw image byte arrays. |

### Returns

| Type    | Description                                     |
| ------- | ----------------------------------------------- |
| `error` | Directory creation or image file write failure. |

### Behavior

* Targets directory `{basePath}/summary/post/{ID}/slide`.
* Creates the directory if it does not already exist.
* Writes each image using its zero-based slice index as the filename.
* Files are stored as `{index}.png`, for example `0.png`, `1.png`, and `2.png`.
* Creates or overwrites image files with file permissions `0644`.
* Stops and returns an error if any image fails to write.

## UpdatePostSlide

Atomically replaces all slide assets associated with a Post ID.

### Signature

```go
UpdatePostSlide(ID int64, images [][]byte) error
```

### Parameters

| Name     | Type       | Description                                                   |
| -------- | ---------- | ------------------------------------------------------------- |
| `ID`     | `int64`    | Unique Post identifier.                                       |
| `images` | `[][]byte` | Complete ordered collection of replacement image byte arrays. |

### Returns

| Type    | Description                                                                             |
| ------- | --------------------------------------------------------------------------------------- |
| `error` | Temporary directory creation, image write, target removal, or directory rename failure. |

### Behavior

* Targets `{basePath}/summary/post/{ID}/slide`.
* Creates a temporary directory within the target's parent directory.
* Writes all replacement images into the temporary directory.
* Uses zero-based numerical filenames such as `0.png`, `1.png`, and `2.png`.
* Removes the existing slide directory.
* Renames the temporary directory to the target directory.
* Cleans up the temporary directory if the operation fails before the swap completes.

The temporary directory is created in the same filesystem partition as the target to support atomic directory renaming.

## DeletePostSummary

Deletes the summary file associated with a Post ID.

### Signature

```go
DeletePostSummary(ID int64) error
```

### Parameters

| Name | Type    | Description             |
| ---- | ------- | ----------------------- |
| `ID` | `int64` | Unique Post identifier. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Targets `{basePath}/summary/post/{ID}/summary.txt`.
* Removes the summary file.
* Returns `nil` if the target does not already exist.

## DeletePostSlide

Deletes all slide assets associated with a Post ID.

### Signature

```go
DeletePostSlide(ID int64) error
```

### Parameters

| Name | Type    | Description             |
| ---- | ------- | ----------------------- |
| `ID` | `int64` | Unique Post identifier. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Targets `{basePath}/summary/post/{ID}/slide`.
* Removes the entire slide directory and its contents.
* Returns `nil` if the target does not already exist.

# Topic Operations

## CreateTopicSummary

Creates or overwrites the summary text for a specific Topic.

### Signature

```go
CreateTopicSummary(ID int64, rssHash string, content string) error
```

### Parameters

| Name      | Type     | Description                         |
| --------- | -------- | ----------------------------------- |
| `ID`      | `int64`  | Unique Topic identifier.            |
| `rssHash` | `string` | RSS hash associated with the Topic. |
| `content` | `string` | Summary text to store.              |

### Returns

| Type    | Description                               |
| ------- | ----------------------------------------- |
| `error` | Directory creation or file write failure. |

### Behavior

* Constructs directory name `{ID}_{rssHash}`.
* Targets `{basePath}/summary/topic/{ID}_{rssHash}`.
* Creates the directory if it does not already exist.
* Writes the summary to `{basePath}/summary/topic/{ID}_{rssHash}/summary.txt`.
* Creates or overwrites the target file with file permissions `0644`.

## UpdateTopicSummary

Atomically updates the summary text for a specific Topic.

### Signature

```go
UpdateTopicSummary(ID int64, rssHash string, content string) error
```

### Parameters

| Name      | Type     | Description                         |
| --------- | -------- | ----------------------------------- |
| `ID`      | `int64`  | Unique Topic identifier.            |
| `rssHash` | `string` | RSS hash associated with the Topic. |
| `content` | `string` | New summary text.                   |

### Returns

| Type    | Description                                                                           |
| ------- | ------------------------------------------------------------------------------------- |
| `error` | Parent directory creation, temporary file, synchronization, close, or rename failure. |

### Behavior

* Targets `{basePath}/summary/topic/{ID}_{rssHash}/summary.txt`.
* Uses atomic temporary-file replacement.
* Synchronizes the temporary file before replacing the target.
* Removes the temporary file if the operation fails.

## CreateTopicSlide

Creates or overwrites the PNG slide assets associated with a Topic.

### Signature

```go
CreateTopicSlide(ID int64, rssHash string, images [][]byte) error
```

### Parameters

| Name      | Type       | Description                                  |
| --------- | ---------- | -------------------------------------------- |
| `ID`      | `int64`    | Unique Topic identifier.                     |
| `rssHash` | `string`   | RSS hash associated with the Topic.          |
| `images`  | `[][]byte` | Ordered collection of raw image byte arrays. |

### Returns

| Type    | Description                                     |
| ------- | ----------------------------------------------- |
| `error` | Directory creation or image file write failure. |

### Behavior

* Constructs directory `{ID}_{rssHash}`.
* Targets `{basePath}/summary/topic/{ID}_{rssHash}/slide`.
* Creates the directory if it does not already exist.
* Writes each image using its zero-based slice index as the filename.
* Files are stored as `{index}.png`.
* Creates or overwrites image files with file permissions `0644`.
* Stops and returns an error if any image fails to write.

## UpdateTopicSlide

Atomically replaces all slide assets associated with a Topic.

### Signature

```go
UpdateTopicSlide(ID int64, rssHash string, images [][]byte) error
```

### Parameters

| Name      | Type       | Description                                                   |
| --------- | ---------- | ------------------------------------------------------------- |
| `ID`      | `int64`    | Unique Topic identifier.                                      |
| `rssHash` | `string`   | RSS hash associated with the Topic.                           |
| `images`  | `[][]byte` | Complete ordered collection of replacement image byte arrays. |

### Returns

| Type    | Description                                                                             |
| ------- | --------------------------------------------------------------------------------------- |
| `error` | Temporary directory creation, image write, target removal, or directory rename failure. |

### Behavior

* Targets `{basePath}/summary/topic/{ID}_{rssHash}/slide`.
* Creates a temporary directory within the target's parent directory.
* Writes all replacement images into the temporary directory.
* Uses zero-based numerical filenames such as `0.png`, `1.png`, and `2.png`.
* Removes the existing slide directory.
* Renames the temporary directory to the target directory.
* Cleans up the temporary directory if the operation fails before the swap completes.

## DeleteTopicSummary

Deletes the summary file associated with a Topic.

### Signature

```go
DeleteTopicSummary(ID int64, rssHash string) error
```

### Parameters

| Name      | Type     | Description                         |
| --------- | -------- | ----------------------------------- |
| `ID`      | `int64`  | Unique Topic identifier.            |
| `rssHash` | `string` | RSS hash associated with the Topic. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Targets `{basePath}/summary/topic/{ID}_{rssHash}/summary.txt`.
* Removes the summary file.
* Returns `nil` if the target does not already exist.

## DeleteTopicSlide

Deletes all slide assets associated with a Topic.

### Signature

```go
DeleteTopicSlide(ID int64, rssHash string) error
```

### Parameters

| Name      | Type     | Description                         |
| --------- | -------- | ----------------------------------- |
| `ID`      | `int64`  | Unique Topic identifier.            |
| `rssHash` | `string` | RSS hash associated with the Topic. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Targets `{basePath}/summary/topic/{ID}_{rssHash}/slide`.
* Removes the entire slide directory and its contents.
* Returns `nil` if the target does not already exist.

# Internal File Operations

## checkDirExist

Ensures that a directory exists before a write operation.

### Signature

```go
checkDirExist(path string) error
```

### Parameters

| Name   | Type     | Description                         |
| ------ | -------- | ----------------------------------- |
| `path` | `string` | Directory path to create or verify. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Directory creation failure. |

### Behavior

* Creates the complete directory hierarchy using `os.MkdirAll`.
* Uses directory permissions `0755`.
* Returns `nil` if the directory already exists.

## removePath

Safely removes a filesystem artifact.

### Signature

```go
removePath(path string) error
```

### Parameters

| Name   | Type     | Description                       |
| ------ | -------- | --------------------------------- |
| `path` | `string` | File or directory path to remove. |

### Returns

| Type    | Description                 |
| ------- | --------------------------- |
| `error` | Filesystem removal failure. |

### Behavior

* Checks whether the target exists.
* Returns `nil` when the target does not exist.
* Removes the target using `os.RemoveAll`.
* Can remove both files and directories.
* Returns a wrapped error if removal fails.

## atomicWrite

Atomically replaces a file by writing its contents to a temporary file and renaming it into place.

### Signature

```go
atomicWrite(path string, content []byte) error
```

### Parameters

| Name      | Type     | Description                  |
| --------- | -------- | ---------------------------- |
| `path`    | `string` | Final destination file path. |
| `content` | `[]byte` | Content to write.            |

### Returns

| Type    | Description                                                                         |
| ------- | ----------------------------------------------------------------------------------- |
| `error` | Parent directory, temporary file, write, synchronization, close, or rename failure. |

### Behavior

* Ensures the target's parent directory exists.
* Creates a temporary file in the same directory as the target.
* Writes the complete content to the temporary file.
* Calls `Sync()` to flush the temporary file.
* Closes the temporary file.
* Renames the temporary file to the target path.
* Removes the temporary file if the operation fails before the rename completes.

The same-directory temporary file ensures that the rename occurs within the same filesystem boundary.

## atomicDirSwap

Atomically replaces a slide asset directory with a newly generated directory.

### Signature

```go
atomicDirSwap(path string, images [][]byte) error
```

### Parameters

| Name     | Type       | Description                                                   |
| -------- | ---------- | ------------------------------------------------------------- |
| `path`   | `string`   | Final destination directory path.                             |
| `images` | `[][]byte` | Complete ordered collection of replacement image byte arrays. |

### Returns

| Type    | Description                                                                            |
| ------- | -------------------------------------------------------------------------------------- |
| `error` | Parent directory, temporary directory, image write, target removal, or rename failure. |

### Behavior

* Ensures the target's parent directory exists.
* Creates a temporary directory inside the target's parent directory.
* Writes all images sequentially into the temporary directory.
* Stores images using zero-based numerical filenames such as `0.png`, `1.png`, and `2.png`.
* Uses file permissions `0600` for temporary image files.
* Removes the existing target directory after all replacement images have been successfully written.
* Renames the temporary directory to the final target path.
* Removes the temporary directory if the operation fails before the swap completes.

The temporary directory is created on the same filesystem as the destination so that the final rename can be performed atomically.
