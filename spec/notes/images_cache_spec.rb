RSpec.describe Notes::ImagesCache do
  subject(:images_cache) { described_class.new }

  include_context "with cleanshot helpers"

  before do
    stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
    stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
  end

  it "downloads an image to local cache" do
    Dir.mktmpdir do |local_images_path|
      allow(Notes::Configuration).to receive(:local_images_path).and_return(local_images_path)
      result = images_cache.get(url: cleanshot_url, scope: "SCOPE")
      expect(File.exist?(File.join(local_images_path, result))).to be_truthy
      expect(File.exist?(File.join(local_images_path, "index.json"))).to be_truthy
      expect(File.extname(result)).to eq(".jpg")
    end
  end
end
