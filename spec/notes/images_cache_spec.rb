RSpec.describe Notes::ImagesCache do
  include_context "cleanshot"

  subject(:images_cache) { described_class.new }

  before do
    stub_cleanshot_url.to_return(status: 302, headers: {"Location" => cleanshot_direct_image_url})
    stub_cleanshot_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
  end

  it "downloads an image to local cache" do
    Dir.mktmpdir do |images_path|
      allow(Notes::Configuration).to receive(:images_path).and_return(images_path)
      result = images_cache.get(cleanshot_url)
      expect(File.exist?(File.join(images_path, result))).to be_truthy
      expect(File.exist?(File.join(images_path, "index.json"))).to be_truthy
      expect(File.extname(result)).to eq(".jpg")
    end
  end
end
