RSpec.describe Notes::MarkdownRenderer do
  subject(:renderer) { described_class.new(with_toc_data: true) }

  include_context "cleanshot"

  let(:result) { Redcarpet::Markdown.new(renderer).render(markdown_source) }

  context "cleanshot image rendering" do
    let(:markdown_source) { "![title](#{cleanshot_url} \"alt text\")" }
    let(:expected_html) { %r{^<p><img src="[\d\w]+.jpg" alt="title" title="alt text"></p>\n$} }

    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
      stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
    end

    it "render image tag" do
      Dir.mktmpdir do |images_path|
        allow(Notes::Configuration).to receive(:images_path).and_return(images_path)
        expect(result).to match(expected_html)
      end
    end
  end
end
