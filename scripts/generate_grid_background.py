#!/usr/bin/env -S uv run
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "pillow>=10.0.0",
# ]
# ///
"""Generate a 2560x1440 grid background for Shine widget positioning.

This creates a visual reference grid with 16x9 cells (160x160 each) labeled
with their coordinates for use in automated integration testing and development.

Usage:
    uv run scripts/generate_grid_background.py
    uv run scripts/generate_grid_background.py -o /path/to/output.png
    uv run scripts/generate_grid_background.py --font-size 28
"""

from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import argparse


# Display configuration
DISPLAY_WIDTH = 2560
DISPLAY_HEIGHT = 1440
GRID_COLS = 16
GRID_ROWS = 9
CELL_WIDTH = DISPLAY_WIDTH // GRID_COLS  # 160
CELL_HEIGHT = DISPLAY_HEIGHT // GRID_ROWS  # 160

# Colors (matching the reference image)
BG_COLOR = "#000000"  # Black background
GRID_COLOR = "#ff6600"  # Orange grid lines
TEXT_COLOR = "#ffffff"  # White text
GRID_LINE_WIDTH = 2

# Subgrid configuration
SUBGRID_DIVISIONS = 8  # Each cell divided into 8x8 subcells
SUBGRID_COLOR = "#ff660020"  # Orange with low opacity (~12%)
SUBGRID_LINE_WIDTH = 1
SUBGRID_CROSS_SIZE = 1  # Length of each arm of the cross (pixels)


def find_font(size: int = 24) -> ImageFont.FreeTypeFont:
    """Find and load a suitable monospace font.

    Tries common monospace fonts in order of preference.
    Prioritizes Iosevka custom fonts from ~/.local/share/fonts.
    """
    font_base = Path("/home/starbased/.local/share/fonts")

    # Priority 1: Iosevka custom fonts (in subdirectories)
    iosevka_paths = [
        font_base / "IosevkaCustom/TTF/IosevkaCustom-Regular.ttf",
        font_base / "IosevkaCustom/TTF-Unhinted/IosevkaCustom-Regular.ttf",
        font_base / "IosevkaCustom/TTF/IosevkaCustom-Extended.ttf",
    ]

    for font_path in iosevka_paths:
        if font_path.exists():
            print(f"Using font: {font_path}")
            return ImageFont.truetype(str(font_path), size)

    # Priority 2: Other user fonts
    font_candidates = [
        font_base / "JetBrainsMonoNerdFont-Regular.ttf",
        font_base / "JetBrainsMono-Regular.ttf",
        # System fonts
        Path("/usr/share/fonts/TTF/JetBrainsMono-Regular.ttf"),
        Path("/usr/share/fonts/jetbrains-mono/JetBrainsMono-Regular.ttf"),
        Path("/usr/share/fonts/TTF/DejaVuSansMono.ttf"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
        Path("/usr/share/fonts/TTF/LiberationMono-Regular.ttf"),
    ]

    for font_path in font_candidates:
        if font_path.exists():
            print(f"Using font: {font_path}")
            return ImageFont.truetype(str(font_path), size)

    # Last resort: search .local/share/fonts directory for any .ttf
    if font_base.exists():
        ttf_files = list(font_base.glob("*.ttf"))
        if ttf_files:
            font_path = ttf_files[0]
            print(f"Using fallback font: {font_path}")
            return ImageFont.truetype(str(font_path), size)

    # Absolute last resort: default font
    print("Warning: No suitable font found, using default")
    return ImageFont.load_default()


def draw_grid(draw: ImageDraw.ImageDraw, width: int, height: int) -> None:
    """Draw the grid lines with double-line center axes (like ═ and ║)."""
    # Center axes positions (between quadrants)
    center_x = 8 * CELL_WIDTH  # Between columns 7 and 8 (at x=1280)
    center_y = 5 * CELL_HEIGHT  # Between rows 4 and 5 (at y=800)

    # Spacing for double lines (distance between parallel lines)
    double_spacing = 3

    # Line offset to ensure consistent appearance at edges
    # When line width is 2, offset by 1 to keep the line fully inside bounds
    edge_offset = GRID_LINE_WIDTH // 2

    # Draw vertical lines
    for col in range(GRID_COLS + 1):
        x = col * CELL_WIDTH

        # Adjust edge positions to ensure consistent line thickness
        if x == 0:
            x = edge_offset
        elif x >= width:
            x = width - 1 - edge_offset

        # Double line for center vertical axis (like ║)
        if col == 8:  # Center column
            draw.line([(x - double_spacing, edge_offset),
                      (x - double_spacing, height - 1 - edge_offset)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)
            draw.line([(x + double_spacing, edge_offset),
                      (x + double_spacing, height - 1 - edge_offset)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)
        else:
            draw.line([(x, edge_offset), (x, height - 1 - edge_offset)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)

    # Draw horizontal lines
    for row in range(GRID_ROWS + 1):
        y = row * CELL_HEIGHT

        # Adjust edge positions to ensure consistent line thickness
        if y == 0:
            y = edge_offset
        elif y >= height:
            y = height - 1 - edge_offset

        # Double line for center horizontal axis (like ═)
        if row == 5:  # Center row
            draw.line([(edge_offset, y - double_spacing),
                      (width - 1 - edge_offset, y - double_spacing)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)
            draw.line([(edge_offset, y + double_spacing),
                      (width - 1 - edge_offset, y + double_spacing)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)
        else:
            draw.line([(edge_offset, y), (width - 1 - edge_offset, y)],
                     fill=GRID_COLOR, width=GRID_LINE_WIDTH)


def draw_subgrid(draw: ImageDraw.ImageDraw, width: int, height: int) -> None:
    """Draw light subgrid crosses at intersection points (like ┼).

    Divides each 160×160 cell into 8×8 subcells of 20×20 pixels each.
    Draws small crosses at each intersection point within cell boundaries.
    """
    subcell_size = CELL_WIDTH // SUBGRID_DIVISIONS  # 20 pixels
    edge_offset = GRID_LINE_WIDTH // 2

    # Draw subgrid within each cell
    for row in range(GRID_ROWS):
        for col in range(GRID_COLS):
            cell_x = col * CELL_WIDTH
            cell_y = row * CELL_HEIGHT

            # Determine cell boundaries (accounting for edge borders)
            cell_x_start = cell_x + edge_offset if col == 0 else cell_x
            cell_y_start = cell_y + edge_offset if row == 0 else cell_y
            cell_x_end = cell_x + CELL_WIDTH - edge_offset if col == GRID_COLS - 1 else cell_x + CELL_WIDTH
            cell_y_end = cell_y + CELL_HEIGHT - edge_offset if row == GRID_ROWS - 1 else cell_y + CELL_HEIGHT

            # Draw crosses at subgrid intersections
            for sub_row in range(1, SUBGRID_DIVISIONS):
                for sub_col in range(1, SUBGRID_DIVISIONS):
                    x = cell_x + sub_col * subcell_size
                    y = cell_y + sub_row * subcell_size

                    # Skip if intersection is outside cell bounds
                    if x < cell_x_start or x >= cell_x_end:
                        continue
                    if y < cell_y_start or y >= cell_y_end:
                        continue

                    # Draw horizontal line of cross
                    x1 = max(x - SUBGRID_CROSS_SIZE, cell_x_start)
                    x2 = min(x + SUBGRID_CROSS_SIZE, cell_x_end - 1)
                    draw.line([(x1, y), (x2, y)],
                             fill=SUBGRID_COLOR, width=SUBGRID_LINE_WIDTH)

                    # Draw vertical line of cross
                    y1 = max(y - SUBGRID_CROSS_SIZE, cell_y_start)
                    y2 = min(y + SUBGRID_CROSS_SIZE, cell_y_end - 1)
                    draw.line([(x, y1), (x, y2)],
                             fill=SUBGRID_COLOR, width=SUBGRID_LINE_WIDTH)


def draw_cell_labels(
    draw: ImageDraw.ImageDraw,
    center_font: ImageFont.FreeTypeFont,
    corner_font: ImageFont.FreeTypeFont,
) -> None:
    """Draw coordinate labels in each cell.

    Shows cell coordinates (col,row) centered and pixel coordinates
    of the top-left corner in the top-left of each cell.
    """
    for row in range(GRID_ROWS):
        for col in range(GRID_COLS):
            # Calculate cell position
            cell_x = col * CELL_WIDTH
            cell_y = row * CELL_HEIGHT

            # Center label: cell coordinates
            center_label = f"{col},{row}"
            bbox = draw.textbbox((0, 0), center_label, font=center_font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]

            # Calculate centered position
            center_x = cell_x + (CELL_WIDTH - text_width) // 2
            center_y = cell_y + (CELL_HEIGHT - text_height) // 2

            # Draw center label
            draw.text((center_x, center_y), center_label, fill=TEXT_COLOR, font=center_font)

            # Top-left corner label: pixel coordinates
            corner_label = f"{cell_x},{cell_y}"

            # Position in top-left corner with small padding
            corner_x = cell_x + 4
            corner_y = cell_y + 4

            # Draw corner label
            draw.text((corner_x, corner_y), corner_label, fill=TEXT_COLOR, font=corner_font)


def generate_grid_background(
    output_path: Path,
    center_font_size: int = 32,
    corner_font_size: int = 10,
    enable_subgrid: bool = False,
) -> None:
    """Generate the grid background image.

    Args:
        output_path: Where to save the output image
        center_font_size: Size of centered cell coordinate text
        corner_font_size: Size of corner pixel coordinate text
        enable_subgrid: Enable 16×16 subgrid within each cell
    """
    # Create image
    img = Image.new("RGB", (DISPLAY_WIDTH, DISPLAY_HEIGHT), BG_COLOR)
    draw = ImageDraw.Draw(img)

    # Load fonts
    center_font = find_font(size=center_font_size)
    corner_font = find_font(size=corner_font_size)

    # Draw components (order matters: subgrid first, then main grid, then labels on top)
    if enable_subgrid:
        draw_subgrid(draw, DISPLAY_WIDTH, DISPLAY_HEIGHT)

    draw_grid(draw, DISPLAY_WIDTH, DISPLAY_HEIGHT)
    draw_cell_labels(draw, center_font, corner_font)

    # Save image
    output_path.parent.mkdir(parents=True, exist_ok=True)
    img.save(output_path, "PNG")
    print(f"✓ Grid background saved to: {output_path}")
    print(f"  Resolution: {DISPLAY_WIDTH}x{DISPLAY_HEIGHT}")
    print(f"  Grid: {GRID_COLS}x{GRID_ROWS} cells ({CELL_WIDTH}x{CELL_HEIGHT}px each)")
    if enable_subgrid:
        subcell_size = CELL_WIDTH // SUBGRID_DIVISIONS
        print(f"  Subgrid: {SUBGRID_DIVISIONS}×{SUBGRID_DIVISIONS} divisions per cell ({subcell_size}px each)")
    print(f"  Center font size: {center_font_size}px, Corner font size: {corner_font_size}px")


def main() -> None:
    """CLI entry point."""
    parser = argparse.ArgumentParser(
        description="Generate grid background for Shine widget positioning"
    )
    parser.add_argument(
        "-o",
        "--output",
        type=Path,
        default=Path("/home/starbased/Pictures/shine-grid-background.png"),
        help="Output file path (default: ~/Pictures/shine-grid-background.png)",
    )
    parser.add_argument(
        "--center-font-size",
        type=int,
        default=32,
        help="Font size for centered cell coordinates (default: 32)",
    )
    parser.add_argument(
        "--corner-font-size",
        type=int,
        default=10,
        help="Font size for corner pixel coordinates (default: 10)",
    )
    parser.add_argument(
        "--subgrid",
        action="store_true",
        help="Enable 16×16 subgrid within each cell (10px per subcell)",
    )

    args = parser.parse_args()

    generate_grid_background(
        output_path=args.output,
        center_font_size=args.center_font_size,
        corner_font_size=args.corner_font_size,
        enable_subgrid=args.subgrid,
    )


if __name__ == "__main__":
    main()
