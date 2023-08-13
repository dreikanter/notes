require_relative "../src/configuration"
require_relative "../src/markdown_renderer"

RSpec.describe MarkdownRenderer do
  subject(:renderer) { described_class.new(with_toc_data: true) }

  let(:markdown_source) do
    <<~MARKDOWN
      # Hello world
      ![title](https://example.com/image.jpg "alt text")
    MARKDOWN
  end

  let(:expected_html) do
    <<~HTML
      <h1 id="hello-world">Hello world</h1>

      <p><img src="https://example.com/image.jpg" alt="title" title="alt text"></p>
    HTML
  end

  it { expect(Redcarpet::Markdown.new(renderer).render(markdown_source)).to eq(expected_html) }
end
