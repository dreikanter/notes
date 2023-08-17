RSpec.describe Notes::MarkdownRenderer do
  subject(:renderer) do
    described_class.new(
      extensions: {with_toc_data: true},
      process_image: process_image
    )
  end

  include_context "with cleanshot helpers"

  let(:result) { Redcarpet::Markdown.new(renderer).render(markdown_source) }
  let(:process_image) { Proc.new {} }

  context "with cleanshot image reference" do
    let(:markdown_source) { "![title](#{cleanshot_url} \"alt text\")" }
    let(:expected_html) { "<p><img src=\"CACHED_IMAGE_PATH\" alt=\"title\" title=\"alt text\"></p>\n" }

    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
      stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
    end

    it "render image tag" do
      Dir.mktmpdir do |images_path|
        allow(Notes::Configuration).to receive(:images_path).and_return(images_path)
        expect(process_image).to receive(:call).once.and_return("CACHED_IMAGE_PATH")
        expect(result).to match(expected_html)
      end
    end
  end
end
