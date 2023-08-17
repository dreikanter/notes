RSpec.describe Notes::MarkdownRenderer do
  subject(:renderer) do
    described_class.new(
      extensions: {with_toc_data: true},
      process_image: process_image
    )
  end

  let(:result) { Redcarpet::Markdown.new(renderer).render(markdown_source) }
  let(:process_image) do
    lambda do |url|
      raise unless url == "https://example.com/image.jpg"
      "PROCESSED_IMAGE_URL"
    end
  end

  context "with cleanshot image reference" do
    let(:markdown_source) { "![title](https://example.com/image.jpg \"alt text\")" }
    let(:expected_html) { "<p><img src=\"PROCESSED_IMAGE_URL\" alt=\"title\" title=\"alt text\"></p>\n" }

    it { expect(result).to match(expected_html) }
  end
end
