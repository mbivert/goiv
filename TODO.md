# Moving lag @bug-move-lagging
Drawing too much.

# Test gfx zooming
See <https://github.com/veandco/go-sdl2/blob/master/gfx/sdl_gfx.go>, e.g. gfx.ZoomSurface()

# Renaming/tagging @renaming @tagging
Use conventions from <https://tales.mbivert.com/on-tagging-files-simply/> for
tagging. See also @exec and @configuration.

# Current filename to copy buffer @copy-fn

# File count display @file-count

# Persistent zoom @persistent-zoom
zoom/clickX/clickY could be stored on a per-image basis

# Shortcuts, configuration @shortcuts @configuration
Shortcuts could be made more generic; have a proper
.json configuration file.

# Execute command on current file @exec
Notably, this could allow to run convert(1) on the current path. Frequent
commands could be registered as @shortcuts. A placeholder (e.g. %i) or
an environment variable could be used to access the current filename within
the command to be run.

Note that this may be generic enough to allow @renaming and @tagging

# Resilient on bad files @resiliency
Do not quit but skip on image loading failure => manage the case
where all images fail.

# Images formats @img-fmt
Explicit list is unfortunate

# Read input from stdin @stdin
Have an option / default to reading a list of paths from
stdin, one per line.

# Better .svg support @svg-zoom

# Better .gif support @gif-support
We probably can animate them; zoom is broken
