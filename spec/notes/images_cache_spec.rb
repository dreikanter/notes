RSpec.describe Notes::ImagesCache do
  subject(:images_cache) { described_class.new }

  include_context "with cleanshot helpers"

  before do
    stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
    stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
  end

  it "downloads an image to local cache" do
    Dir.mktmpdir do |assets_path|
      allow(Notes::Configuration).to receive(:assets_path).and_return(assets_path)
      result = images_cache.get(url: cleanshot_url, scope: "SCOPE")
      expect(File).to exist(File.join(assets_path, result))
      expect(File).to exist(File.join(assets_path, "index.json"))
      expect(File.extname(result)).to eq(".jpg")
    end
  end
end
