#simple util
#fully coded by gemini

import os
import struct
from PIL import Image

def png_to_rgb565_chromakey(png_path, bin_path, alpha_threshold=128):
    """
    Converts a PNG with transparency into a raw 16-bit RGB565 binary file.
    Transparent pixels are automatically replaced with Pure Magenta (0xF81F).
    """
    # Open image and force it to RGBA mode to ensure an alpha channel exists
    img = Image.open(png_path).convert('RGBA')
    width, height = img.size
    
    # 16-bit value for pure magenta (R=31, G=0, B=31) -> 0xF81F
    # In Little-Endian binary format, this gets written as bytes: \x1f\xf8
    chroma_key_16bit = (31 << 11) | (0 << 5) | 31 
    
    with open(bin_path, 'wb') as f:
        for y in range(height):
            for x in range(width):
                r, g, b, a = img.getpixel((x, y))
                
                # If the pixel is mostly transparent, force it to our ChromaKey color
                if a < alpha_threshold:
                    # Write directly as a 2-byte little-endian integer ('<H')
                    f.write(struct.pack('<H', chroma_key_16bit))
                else:
                    # Scale down standard 8-bit color channels (0-255) to 16-bit
                    r5 = (r >> 3) & 0x1F  # 5 bits Red
                    g6 = (g >> 2) & 0x3F  # 6 bits Green
                    b5 = (b >> 3) & 0x1F  # 5 bits Blue
                    
                    # Pack channels into a single uint16 integer
                    rgb565 = (r5 << 11) | (g6 << 5) | b5
                    
                    # Safeguard: If the actual artwork accidentally uses pure magenta,
                    # nudge it slightly so it doesn't get rendered as invisible.
                    if rgb565 == chroma_key_16bit:
                        rgb565 = (r5 << 11) | (1 << 5) | b5 # Add 1 unit of green
                        
                    f.write(struct.pack('<H', rgb565))
                    
    print(f"Successfully converted: {png_path} -> {bin_path} ({width}x{height})")

if __name__ == "__main__":
    # Example 1: Convert a single sprite file
    # Change "circle.png" to the name of your graphic file
    input_file = "image.png"
    output_file = "image.bin"
    
    if os.path.exists(input_file):
        png_to_rgb565_chromakey(input_file, output_file)
    else:
        print(f"Error: Could not find '{input_file}' in this directory.")
        print("Please place a PNG image in this folder or edit the script paths.")
