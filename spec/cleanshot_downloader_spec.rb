RSpec.describe Notes::CleanshotDownloader do
  subject(:service_call) { described_class.new(url).download }

  let(:url) { "https://share.cleanshot.com/80085" }
  let(:image_content) { file_fixture("banana.jpg").read }
  let(:original_file_name) { "CleanShot 2023-08-06 at 15.27.35.jpeg" }

  let(:direct_image_url) do
    "https://media.cleanshot.cloud/media/1334/itkYgfyMGKeaUT.jpeg?" \
    "response-content-disposition=attachment%3Bfilename%3DCleanShot" \
    "%25202023-08-06%2520at%252015.27.35.jpeg&Expires=1691367016"
  end

  context "with successful responses" do
    before do
      stub_url.to_return(status: 302, headers: {"Location" => direct_image_url})
      stub_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_content)
    end

    it { expect(service_call).to eq(url: url, file_name: original_file_name, content: image_content) }
  end

  context "with redirect error" do
    before do
      stub_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error getting image URL") }
  end

  context "with image downloading error" do
    before do
      stub_url.to_return(status: 302, headers: {"Location" => direct_image_url})
      stub_direct_image_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error downloading image") }
  end

  def stub_url
    stub_request(:get, "#{url}+")
  end

  def stub_direct_image_url
    stub_request(:get, direct_image_url)
  end
end
