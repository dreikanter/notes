RSpec.describe Notes::CleanshotDownloader do
  include_context "cleanshot"

  subject(:service_call) { described_class.new(cleanshot_url).download }

  context "with successful responses" do
    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
      stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
    end

    it "cache the image" do
      Dir.mktmpdir do |images_path|
        allow(Notes::Configuration).to receive(:images_path).and_return(images_path)
        expect(service_call).to eq(url: cleanshot_url, file_name: cleanshot_original_file_name, content: image_content)
      end
    end
  end

  context "with redirect error" do
    before do
      stub_cleanshot_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error getting image URL") }
  end

  context "with image downloading error" do
    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
      stub_cleanshot_direct_image_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error downloading image") }
  end
end
