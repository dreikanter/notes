require_relative "../src/markdown_renderer"

RSpec.describe MarkdownRenderer do
  subject(:renderer) { described_class.new }

  it { expect(renderer.image("URL", "ALT", "TITLE")).to eq("<img src=\"URL\" alt=\"TITLE\" title=\"ALT\">") }
end
