/* -*- coding: utf-8 -*-
 * vim: filetype=cpp
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Debugger - Memory Editor Window
 */
#pragma once

// standard library headers
#include <cstdint>
#include <cstdio>
#include <cstring>
#include <algorithm>
#include <string>
#include <imgui.h>

// program-specific includes
#include "generic.h"
#include "../constants.h"
#include "../debug_state.h"


// class definition
class UIMemoryView : public IGeneric
{
public:
    UIMemoryView(DebugState& state, GLFWwindow* window) :
        IGeneric(state, window, MEMORYVIEW_HWND)
    {
        mem_data_ = state.memview<uint8_t>(0, ram_size).data();
        mem_size_ = ram_size;
    }

    ~UIMemoryView()
    { }


    virtual void render()
    {
        if (!state_.isVisible(id_)) return;

        ImGui::SetNextWindowSize(ImVec2(580,400), ImGuiCond_FirstUseEver);
        if (!ImGui::Begin("Memory Viewer", &state_.get(id_))) {
            ImGui::End();
            return;
        }

        // effective region
        uint32_t eff_start = region_start_;
        uint32_t eff_size = (region_size_ > 0) ? region_size_ : static_cast<uint32_t>(mem_size_);
        uint32_t eff_end = eff_start + eff_size;

        // top-bar: go-to address
        drawTopBar(eff_start, eff_end);
        ImGui::Separator();

        // hex view: scrollable window
        drawHexView(eff_start, eff_size);
        ImGui::Separator();

        // bottom-bar: region info and selection
        drawBottomBar(eff_start, eff_end);

        ImGui::End();
    }

private:
    //----- members
    int columns_ = 16;                      // bytes per row
    bool show_ascii_ = true;                // show ASCII panel
    bool show_options_ = false;             // show options panel
    bool uppercase_hex_ = true;             // uppercase hex digits

    uint8_t* mem_data_ = nullptr;
    size_t mem_size_ = 0;

    // --- state
    bool visible_ = true;                   // window visibility
    uint32_t base_addr_ = 0x0000'0000;           // display base address
    uint32_t region_start_ = 0x0000'0000;        // viewable region start
    uint32_t region_size_ = 0;                   // viewable region size (0 = full RAM)
    int selected_addr_ = -1;               // currently selected byte (-1 = none)
    char addr_input_[9] = "00000000";       // go-to address input buffer

    // --- colors
    ImU32 col_address_    = IM_COL32(0x80, 0x80, 0x80, 0xFF);  // grey addresses
    ImU32 col_hex_zero_   = IM_COL32(0x60, 0x60, 0x60, 0xFF);  // dimmed zero bytes
    ImU32 col_hex_normal_ = IM_COL32(0xE0, 0xE0, 0xE0, 0xFF);  // normal byte value
    ImU32 col_ascii_dot_  = IM_COL32(0x50, 0x50, 0x50, 0xFF);  // non-printable ASCII
    ImU32 col_ascii_char_ = IM_COL32(0xA0, 0xD0, 0xA0, 0xFF);  // printable ASCII
    ImU32 col_highlight_  = IM_COL32(0x26, 0x4F, 0x78, 0xFF);  // selection highlight
    ImU32 col_hover_      = IM_COL32(0x26, 0x4F, 0x78, 0xFF);  // selection highlight
    ImU32 col_modified_   = IM_COL32(0xFF, 0x60, 0x60, 0xFF);  // modified byte


    //----- methods
    void drawTopBar(uint32_t eff_start, uint32_t eff_end)
    {
        ImGui::SetNextItemWidth(70);
        if (ImGui::InputText("##addr", addr_input_, sizeof(addr_input_),
                             ImGuiInputTextFlags_CharsHexadecimal |
                             ImGuiInputTextFlags_EnterReturnsTrue))
        {
            uint32_t target = 0;
            if (std::sscanf(addr_input_, "%X", &target) == 1) {
                if (target >= eff_start && target < eff_end) {
                    // align to row and scroll here
                    base_addr_ = target & ~(static_cast<uint32_t>(columns_) - 1);
                    selected_addr_ = static_cast<int>(target);
                }
            }
        }

        ImGui::SameLine();
        if (ImGui::Button("Go!")) {
            uint32_t target = 0;
            if (std::sscanf(addr_input_, "%X", &target) == 1) {
                if (target >= eff_start && target < eff_end) {
                    // align to row and scroll here
                    base_addr_ = target & ~(static_cast<uint32_t>(columns_) - 1);
                    selected_addr_ = static_cast<int>(target);
                }
            }
        }

        // display format for the selected byte
        ImGui::SameLine();
        ImGui::Spacing();
        ImGui::SameLine();
        if (selected_addr_ >= 0 && selected_addr_ < static_cast<int>(mem_size_)) {
            uint8_t byte_val = mem_data_[selected_addr_];
            // show as multiple representations
            ImGui::Text("Dec: %3d (Uint8)", byte_val);
            ImGui::SameLine();
            ImGui::Spacing();

            ImGui::SameLine();
            ImGui::Text("Hex: %02X", byte_val);
            ImGui::SameLine();
            ImGui::Spacing();

            ImGui::SameLine();
            ImGui::Text("Bin: %d%d%d%d %d%d%d%d",
                (byte_val >> 7) & 1, (byte_val >> 6) & 1,
                (byte_val >> 5) & 1, (byte_val >> 4) & 1,
                (byte_val >> 3) & 1, (byte_val >> 2) & 1,
                (byte_val >> 1) & 1, (byte_val >> 0) & 1);
        }
    }

    void drawHexView(uint32_t eff_start, uint32_t eff_size)
    {
        // compute layout metrics
        ImGuiStyle& style = ImGui::GetStyle();
        float char_width = ImGui::CalcTextSize("F").x;
        float glyph_width = ImGui::CalcTextSize("F").x;
        float cell_width = ImGui::CalcTextSize("FF").x + 0.5 * glyph_width;

        // column header
        ImGui::PushStyleColor(ImGuiCol_Text, col_address_);
        ImGui::Text("ADDR");
        for (int col = 0; col < columns_; col++) {
            ImGui::SameLine(getHexCellPosX(col, char_width));
            ImGui::Text("%02X", col);
        }
        if (show_ascii_) {
            ImGui::SameLine(getAsciiPosX(char_width));
            ImGui::Text("ASCII");
        }
        ImGui::PopStyleColor();

        ImGui::Separator();

        // scrollable region
        uint32_t total_rows = (eff_size + columns_ - 1) / columns_;
        float line_height = ImGui::GetTextLineHeightWithSpacing();
        float footer_height = ImGui::GetFrameHeightWithSpacing() + ImGui::GetStyle().ItemSpacing.y;
        ImGui::BeginChild("##hex_scroll", ImVec2(0, -footer_height), ImGuiChildFlags_None,
                          ImGuiWindowFlags_NoMove);

        // use clipper for large memory regions
        ImGuiListClipper clipper;
        clipper.Begin(static_cast<int>(total_rows), line_height);

        while (clipper.Step()) {
            for (int row = clipper.DisplayStart; row < clipper.DisplayEnd; row++) {
                uint32_t row_addr = eff_start + static_cast<uint32_t>(row * columns_);
                ImDrawList* draw_list = ImGui::GetWindowDrawList();

                // address column
                ImGui::PushStyleColor(ImGuiCol_Text, col_address_);
                ImGui::Text("%04X:", row_addr);
                ImGui::PopStyleColor();

                // hex bytes
                for (int col = 0; col < columns_; col++) {
                    uint32_t addr = row_addr + col;
                    if (addr >= eff_start + eff_size) break;

                    float pos_x = getHexCellPosX(col, char_width);
                    ImGui::SameLine(pos_x);

                    uint8_t byte_val = mem_data_[addr];

                    // color: dimmed for zero, normal otherwise
                    ImU32 byte_color = (byte_val == 0) ? col_hex_zero_ : col_hex_normal_;
                    ImGui::PushStyleColor(ImGuiCol_Text, byte_color);

                    // clickable hex cell
                    char id_buf[16];
                    std::snprintf(id_buf, sizeof(id_buf), "##%08X", addr);
                    ImGui::PushID(static_cast<int>(addr));

                    ImGui::PushStyleColor(ImGuiCol_HeaderHovered, IM_COL32(0, 0, 0, 0));
                    ImGui::PushStyleColor(ImGuiCol_HeaderActive, IM_COL32(0, 0, 0, 0));

                    if (ImGui::Selectable(id_buf, false, 0,
                                          ImVec2(char_width*0.7*2, line_height)))     // ImVec2(cell_width - char_width * 0.5f, line_height)))
                    {
                        selected_addr_ = static_cast<int>(addr);
                        std::snprintf(addr_input_, sizeof(addr_input_), "%08X", addr);
                    }

                    ImGui::PopStyleColor(2);

                    // everything below uses the ACTUAL rect of the selectable we just drew
                    ImVec2 item_min = ImGui::GetItemRectMin();
                    ImVec2 item_max = ImGui::GetItemRectMax();

                    item_min.x -= (int)(glyph_width / 2);
                    item_max.x -= (int)(glyph_width / 2);

                    // highlight selected byte
                    if (static_cast<int>(addr) == selected_addr_) {
                        draw_list->AddRectFilled(item_min, item_max, col_highlight_);
                    } else if (ImGui::IsItemHovered()) {
                        draw_list->AddRectFilled(item_min, item_max, col_hover_);
                    }

                    // overlay the hex text on top of the selectable
                    ImVec2 text_pos = ImGui::GetItemRectMin();
                    const char* fmt = uppercase_hex_ ? "%02X" : "%02x";
                    char hex_buf[4];
                    std::snprintf(hex_buf, sizeof(hex_buf), fmt, byte_val);
                    draw_list->AddText(text_pos, byte_color, hex_buf);
                    ImGui::PopID();

                    ImGui::PopStyleColor();
                }

                // ASCII column
                if (show_ascii_) {
                    ImGui::SameLine(getAsciiPosX(char_width));

                    for (int col = 0; col < columns_; col++) {
                        uint32_t addr = row_addr + col;
                        if (addr >= eff_start + eff_size) break;

                        uint8_t byte_val = mem_data_[addr];
                        bool printable = (byte_val >= 0x20 && byte_val < 0x7F);
                        char c = printable ? static_cast<char>(byte_val) : '.';
                        ImU32 ascii_col = printable ? col_ascii_char_ : col_ascii_dot_;

                        // highlight in ASCII panel too
                        if (static_cast<int>(addr) == selected_addr_) {
                            ImVec2 pos = ImGui::GetCursorScreenPos();
                            draw_list->AddRectFilled(
                                pos,
                                ImVec2(pos.x + glyph_width, pos.y + line_height),
                                col_highlight_
                            );
                        }

                        char ascii_buf[2] = { c, '\0' };
                        ImVec2 pos = ImGui::GetCursorScreenPos();
                        draw_list->AddText(pos, ascii_col, ascii_buf);
                        ImGui::SetCursorScreenPos(ImVec2(pos.x + glyph_width, pos.y));
                    }
                    ImGui::NewLine();
                }
            }
        }
        clipper.End();

        ImGui::EndChild();
    }

    void drawBottomBar(uint32_t eff_start, uint32_t eff_end)
    {
        ImGui::PushStyleColor(ImGuiCol_Text, col_address_);
        ImGui::Text("REGION: %04X-%04X", eff_start, eff_end - 1);
        ImGui::SameLine();
        if (selected_addr_ >= 0) {
            ImGui::Text("  SELECTION: %08X", static_cast<uint32_t>(selected_addr_));
        }
        ImGui::PopStyleColor();
    }

    // compute the X position for a hex cell at the given column
    float getHexCellPosX(int col, float char_width) const
    {
        // "XXXX: " address = ~6 chars, then each byte = 3 chars
        // extra gap every 8 bytes
        float base = char_width * 7.0f;         // after "XXXX: "
        float cell = char_width * 3.0f;         // "XX " per byte
        float gap  = char_width * 1.0f;         // extra gap every 8 bytes

        float x = base + col * cell;
        x += (col / 8) * gap;                   // group separator
        return x;
    }

    // compute the X position for the ASCII column
    float getAsciiPosX(float char_width) const
    {
        float hex_end = getHexCellPosX(columns_, char_width);
        return hex_end + char_width * 2.0f;      // small gap before ASCII
    }
};
